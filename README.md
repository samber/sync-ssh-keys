# Github teams ssh keys

Sync public ssh keys to `~/.ssh/authorized_keys`, based on Github teams membership.

## Install

```
go get github.com/samber/github-team-ssh-keys
```

or

```
curl -o /usr/local/bin/github-team-ssh-keys https://github.com/samber/github-team-ssh-keys/releases/download/v0.1.0/github-team-ssh-keys_v0.1.0_linux-amd64
```

### Sync with crontask

```
crontab -e
```

Then:

```
# sync once per hour
0 * * * * github-team-ssh-keys --github-token XXXXXXXXXXXXXXX --github-org epitech --github-team sysadmin > /root/.ssh/authorized_keys
```

## Usage



## Contribute

## License

[MIT license](./LICENSE)
