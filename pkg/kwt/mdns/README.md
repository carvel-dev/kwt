## mDNS on OS X

Useful commands:

- `syslog` shows Apple System Logs (asl)
- `scutil --dns` shows current DNS configuration

### Snippet of discoveryd logs

```bash
$ curl -vvv test.local
```

```bash
$ syslog -k Time ge "Sep 13 12:25:27"
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Intermediate ClientIPC DNSServiceQueryRecord: ID 418 flags=0x15000, interface=0, name=test.local, rrtype=A, rrclass=1 (curl(21264))
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets UDS FD=42 SendReturnStatus=0 bytecnt=4 errno=0
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed ClientIPC Connection 202 closing error return socket FD=42
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets UDS FD=42 close
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets UDS FD=41 Recv bytecnt=53 errno=0
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets UDS FD=41 Recv bytecnt=1 errno=0
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Intermediate Bonjour set up question for test.local
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets UDS FD=42 ErrorReturnSocket
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed ClientIPC Request 202-301 Add event 0x0 en0 : test.local A Record 127.0.0.1, class IN, TTL 65
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Detailed Sockets curl(21264) UDS FD=42
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Intermediate ClientIPC DNSServiceQueryRecord: ID 419 flags=0x15000, interface=0, name=test.local, rrtype=AAAA, rrclass=1 (curl(21264))
...
Sep 13 12:25:27 idora discoveryd[20637] <Notice>: Intermediate Bonjour set up question for test.local
Sep 13 12:25:32 idora discoveryd[20637] <Notice>: Detailed Bonjour timeout Q 418 with flag
```
