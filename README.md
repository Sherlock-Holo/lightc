# lightc
a lightweight container without daemon

## Why I write this project
1. I want to learn container
2. docker is cool, but
3. docker is heavy with daemon

## What features lightc has
- non-big-daemon
- no user-land proxy
- replace iptables with nftables (after stable version)

## what the different between docker and lightc
- lightc is a naive project, you can't run with your company server unless you are crazy :-)
- lightc `run` doesn't have `-i` `-t` flags, only have `--tty` flags