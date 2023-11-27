# uncork

[Corkscrew](https://github.com/bryanpkc/corkscrew) port to Go: a tool for tunneling SSH through HTTP proxies.

```
uncork 0.1.0 (martin.czygan@gmail.com)
usage: uncork <proxyhost> <proxyport> <desthost> <destport>
```

Depending on your Linux distribution, you can also just use:

```
Host github.com
  User git
  ProxyCommand /bin/bash -c 'exec 3<>/dev/tcp/$PROXY_IP/$PROXY_PORT; printf "CONNECT %h:%p HTTP/1.1\n\n" >&3; cat <&3 & : ; exec cat >&3'
```
