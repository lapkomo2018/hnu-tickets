import os,sys,tempfile,subprocess,pathlib,json,time,base64,urllib.request,urllib.error
from cryptography.fernet import Fernet
ROOT=pathlib.Path(__file__).resolve().parent;REPO=ROOT.parent.parent
PG=pathlib.Path(os.environ.get('PG_BIN', '/usr/lib/postgresql/17/bin')).resolve()
API=pathlib.Path(os.environ.get('API_BIN', REPO/'bin/api')).resolve()
CRYPTO=pathlib.Path(os.environ.get('CRYPTO_BIN', REPO/'bin/cryptodata')).resolve()
env=os.environ.copy();env.update(PATH=str(PG)+os.pathsep+env['PATH'],PGHOST='127.0.0.1',PGPORT='15488',PGUSER='agent',PGDATABASE='postgres',POSTGRES_HOST='127.0.0.1',POSTGRES_PORT='15488',POSTGRES_USER='agent',POSTGRES_PASSWORD='local',POSTGRES_DB='drill_source',POSTGRES_SSLMODE='disable',JWT_SECRET=base64.b64encode(os.urandom(32)).decode(),DATA_ACTIVE_KEY='drill',DATA_KEYS=json.dumps({'drill':base64.b64encode(os.urandom(32)).decode()}),BACKUP_KEY=Fernet.generate_key().decode(),APP_PORT='18081',APP_ENV='production')

def run(*args):
 r=subprocess.run(args,env=env,capture_output=True,text=True)
 if r.returncode:
  raise RuntimeError(r.stderr)
 return r.stdout
opener=urllib.request.build_opener(urllib.request.ProxyHandler({}))
def http(path,body=None,token=None):
 headers={'Content-Type':'application/json'}
 if token:headers['Authorization']='Bearer '+token
 req=urllib.request.Request('http://127.0.0.1:18081'+path,data=None if body is None else json.dumps(body).encode(),headers=headers)
 with opener.open(req,timeout=5) as r:return json.load(r)
def start(log):
 p=subprocess.Popen([str(API)],env=env,stdout=log,stderr=log)
 for i in range(100):
  try:http('/ready');return p
  except (OSError,ValueError):time.sleep(.1)
 raise RuntimeError('API not ready')
with tempfile.TemporaryDirectory(prefix='tickets-recovery-') as td:
 d=pathlib.Path(td);run(str(PG/'initdb'),'-D',str(d/'pg'),'-U','agent','-A','trust','--no-locale');run(str(PG/'pg_ctl'),'-D',str(d/'pg'),'-l',str(d/'postgres.log'),'-o',"-h 127.0.0.1 -p 15488 -c unix_socket_directories=''",'-w','start')
 api=None
 try:
  run('createdb','drill_source')
  for p in sorted((REPO/'migrations').glob('*.up.sql')):
   # PostgreSQL 13+ provides gen_random_uuid natively; no other pgcrypto function is used.
   sql=p.read_text().replace('CREATE EXTENSION IF NOT EXISTS pgcrypto;', '')
   run('psql','-v','ON_ERROR_STOP=1','-d','drill_source','-c',sql)
  run(str(CRYPTO),'-mode','migrate')
  log=open(d/'api.jsonl','w');api=start(log)
  auth=http('/api/v1/auth/register',{'name':'Recovery Student','email':'drill@example.test','password':'Recovery-Test-2026'})
  # Handler envelopes use a data field.
  auth=auth.get('data',auth);token=auth['accessToken']
  cat=run('psql','-At','-d','drill_source','-c',"INSERT INTO categories(name,description) VALUES ('Drill','Synthetic') RETURNING id").splitlines()[0]
  ticket=http('/api/v1/tickets',{'title':'Recovery exercise','description':'Synthetic encrypted recovery payload','categoryId':cat,'priority':'MEDIUM'},token);ticket=ticket.get('data',ticket);tid=ticket['id']
  before=http('/api/v1/tickets/'+tid,token=token)
  api.terminate();api.wait();api=None
  backup=d/'backup.enc';ops=str(REPO/'scripts/ops/backup.py')
  saved=json.loads(run(sys.executable,ops,'backup',str(backup),'--database','drill_source'))
  run(sys.executable,ops,'verify',str(backup))
  bad=d/'bad.enc';b=bytearray(backup.read_bytes());b[-10]^=1;bad.write_bytes(b)
  badrun=subprocess.run([sys.executable,ops,'verify',str(bad)],env=env,capture_output=True);assert badrun.returncode!=0
  run('dropdb','drill_source');started=time.monotonic()
  run(sys.executable,ops,'restore',str(backup),'--database','drill_restored');env['POSTGRES_DB']='drill_restored';api=start(log)
  after=http('/api/v1/tickets/'+tid,token=token);assert before==after
  login=http('/api/v1/auth/login',{'email':'drill@example.test','password':'Recovery-Test-2026'});assert login
  result={'backup':saved,'tampered_copy_rejected':True,'source_database_dropped':True,'restored_database':'drill_restored','ticket_equal':True,'login_after_restore':True,'ready_after_restore':True,'recovery_seconds':round(time.monotonic()-started,3)}
  log.flush()
  log_text=(d/'api.jsonl').read_text()
  assert token not in log_text and 'Recovery-Test-2026' not in log_text and 'drill@example.test' not in log_text
  result['secrets_absent_from_log']=True
  (ROOT/'drill.jsonl').write_text(log_text)
  print(json.dumps(result,indent=2));(REPO/'scripts/ops/drill_results.json').write_text(json.dumps(result,indent=2)+'\n')
 finally:
  if api:api.terminate();api.wait()
  run(str(PG/'pg_ctl'),'-D',str(d/'pg'),'-m','immediate','-w','stop')
