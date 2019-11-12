# Github teams ssh keys

Sync public ssh keys to `~/.ssh/authorized_keys`, based on Github teams membership.

## Install

```bash
$ go get github.com/samber/github-team-ssh-keys
```

or

```bash
$ curl -o /usr/local/bin/github-team-ssh-keys \
      https://github.com/samber/github-team-ssh-keys/releases/download/v0.2.0/github-team-ssh-keys_v0.2.0_linux-amd64
$ chmod +x /usr/local/bin/github-team-ssh-keys
```

### Sync using a crontask

```bash
$ crontab -e
```

Then:

```
# sync once per hour
0 * * * * github-team-ssh-keys --github-token XXXXXXXXXXXXXXX --github-org epitech --github-team sysadmin > /root/.ssh/authorized_keys
```

## Usage

### Simple user

```bash
$ github-team-ssh-keys --github-user samber

ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDhDlAK8ewcwCTOv6xZHAAJK36QJ1ljJyn9/GiDTHE9aAREQdTtpPGrLvCxuqy3SZl/hvwSpNFjz0YH0sYvfQvBOTCogNo9o1FKcJaA9jOxPktRb2pObDA0+e2KIbyx3JR4hg63uP+p7awP8uKoRE+O8G6aTmv33mwqsl8ZOMVPo+qEkWniVCc5m7U1a/jIZj2JgFBa7Dhjnnr7RKlUWnmc0VhKQLwiOnyzpSMV2WBlOBrBnUAz60F2exTdX7zgULMHxyRSmL4xe/+BUHEUf9T41AEdWtcUx0iS7m/wGUvHKKokkz1zCkUGFy+Kq3rviH9dWYYt4KiHPm2/6DgKNua/ samber@github-team-ssh-key
```

### All members of an organizations

```bash
$ github-team-ssh-keys --github-token XXXXXXXXXXXXXXX \
                       --github-org epitech

[...]
```

### All members of an organizations being part of teams "root" and "sre"

```bash
$ github-team-ssh-keys --github-token XXXXXXXXXXXXXXX \
                       --github-org epitech \
                       --github-team root \
                       --github-team sre

[...]
```

### All members of an organizations excluding me ;)

```bash
$ github-team-ssh-keys --github-token XXXXXXXXXXXXXXX \
                       --github-org epitech \
                       --exclude-github-user samber

[...]
```

## Contribute

```bash
$ make run
```

```bash
$ make release
```

## License

[MIT license](./LICENSE)
