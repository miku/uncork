# uncork

Go port of [Corkscrew](https://github.com/bryanpkc/corkscrew): a tool for
tunneling SSH through HTTP proxies. This is a usable exercise, but nothing in
excess of what `nc -X ...` could do.

```
uncork 0.1.0 (martin.czygan@gmail.com)
usage: uncork <proxyhost> <proxyport> <desthost> <destport>
```

Put it in your ssh config:

```
Host github.com
  User git
  ProxyCommand uncork proxy.mycompany.com 3128 github.com 22
```

Alternative, via [/dev/tcp](https://tldp.org/LDP/abs/html/devref1.html):

```
Host github.com
  User git
  ProxyCommand /bin/bash -c 'exec 3<>/dev/tcp/$PROXY_IP/$PROXY_PORT; \
        printf "CONNECT %h:%p HTTP/1.1\n\n" >&3; \
        cat <&3 & : ; exec cat >&3'
```

Or just [netcat](https://linux.die.net/man/1/nc):

```
Host github.com
  User git
  ProxyCommand nc -X connect -x proxy_ip:port %h %p
```
