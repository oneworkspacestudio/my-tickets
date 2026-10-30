"""Inspect normal and delay-loaded PE imports without external Python packages."""
import argparse
import hashlib
import json
import pathlib
import struct


def inspect(path):
    data = path.read_bytes()
    if data[:2] != b"MZ":
        raise ValueError("Not a PE executable")
    pe = struct.unpack_from("<I", data, 0x3C)[0]
    if data[pe:pe + 4] != b"PE\0\0":
        raise ValueError("Invalid PE signature")
    machine, sections = struct.unpack_from("<HH", data, pe + 4)
    optional_size = struct.unpack_from("<H", data, pe + 20)[0]
    optional = pe + 24
    magic = struct.unpack_from("<H", data, optional)[0]
    if magic not in (0x10B, 0x20B):
        raise ValueError("Unsupported PE optional header")
    directories = optional + (112 if magic == 0x20B else 96)
    image_base = struct.unpack_from("<Q" if magic == 0x20B else "<I", data, optional + (24 if magic == 0x20B else 28))[0]
    headers_size = struct.unpack_from("<I", data, optional + 60)[0]
    section_table = optional + optional_size

    def offset(rva):
        if rva < headers_size:
            return rva
        for index in range(sections):
            virtual_size, address, raw_size, raw = struct.unpack_from("<IIII", data, section_table + index * 40 + 8)
            if address <= rva < address + max(virtual_size, raw_size):
                result = raw + rva - address
                if result >= len(data):
                    break
                return result
        raise ValueError(f"Unmapped RVA: {rva:x}")

    def name(rva):
        start = offset(rva)
        end = data.find(b"\0", start, min(start + 4096, len(data)))
        if end < 0:
            raise ValueError("Unterminated import name")
        return data[start:end].decode("ascii")

    result = {"file": path.name, "sha256": hashlib.sha256(data).hexdigest(), "machine": hex(machine), "imports": [], "delayImports": []}
    for directory_index, width, target in ((1, 20, "imports"), (13, 32, "delayImports")):
        rva, size = struct.unpack_from("<II", data, directories + directory_index * 8)
        if not rva:
            continue
        start = offset(rva)
        for index in range(min(size // width + 1, 4096)):
            fields = struct.unpack_from("<" + "I" * (width // 4), data, start + index * width)
            if not any(fields):
                break
            name_rva = fields[3] if directory_index == 1 else fields[1]
            if directory_index == 13 and not fields[0] & 1:
                name_rva -= image_base
            result[target].append(name(name_rva))
        else:
            raise ValueError("Import directory has no terminator")
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("executable", type=pathlib.Path)
    args = parser.parse_args()
    print(json.dumps(inspect(args.executable), indent=2))
