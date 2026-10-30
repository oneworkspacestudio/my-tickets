"""Collect pinned dependency attribution without disclosing build-machine paths."""
import argparse,json,shutil,re
from pathlib import Path

def main():
 p=argparse.ArgumentParser();p.add_argument('--go-modules',required=True);p.add_argument('--node-modules',required=True);p.add_argument('--out',required=True);a=p.parse_args()
 out=Path(a.out);out.mkdir(parents=True,exist_ok=True);rows=[]
 def collect(kind,name,version,root,license_id='see licence text'):
  slug=re.sub(r'[^A-Za-z0-9_.-]','_',name+'@'+version)
  texts=[]
  for f in Path(root).iterdir():
   if f.is_file() and re.match(r'^(licen[cs]e|copying|notice|copyright)([.\-_]|$)',f.name,re.I) and f.stat().st_size<4*1024*1024:
    target=out/kind/slug/f.name;target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,target);texts.append(target.relative_to(out).as_posix())
  rows.append({'ecosystem':kind,'name':name,'version':version,'license':license_id,'notices':texts})
 raw=Path(a.go_modules).read_text(encoding='utf-8-sig');decoder=json.JSONDecoder()
 while raw.strip():
  raw=raw.lstrip();m,n=decoder.raw_decode(raw);raw=raw[n:]
  if not m.get('Main') and m.get('Dir'):collect('go',m['Path'],m.get('Version',''),m['Dir'])
 seen=set()
 for package in Path(a.node_modules).glob('.pnpm/*/node_modules/*/package.json'):
  data=json.loads(package.read_text(encoding='utf-8-sig'));key=(data.get('name'),data.get('version'))
  if key in seen:continue
  seen.add(key);collect('npm',key[0],key[1],package.parent,data.get('license','see licence text'))
 for package in Path(a.node_modules).glob('.pnpm/*/node_modules/@*/*/package.json'):
  data=json.loads(package.read_text(encoding='utf-8-sig'));key=(data.get('name'),data.get('version'))
  if key in seen:continue
  seen.add(key);collect('npm',key[0],key[1],package.parent,data.get('license','see licence text'))
 (out/'DEPENDENCIES.json').write_text(json.dumps({'schema':1,'dependencies':sorted(rows,key=lambda x:(x['ecosystem'],x['name'],x['version']))},indent=2),encoding='utf-8')
 print('Collected dependency records:',len(rows),'; licence texts:',len(list(out.rglob('*.*')))-1)
if __name__=='__main__':main()
