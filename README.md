# ssh-cmds
This is similar to pssh, but it's more simple and it returns CSV.
CSV is a human readable, but also a machine readable format.
The benfit of a machine readable format is that you can easily post
process the information.

## Example

    cat hosts | ./ssh-cmds -cmd "dpkg -l nginx"  | tee -a nginx-version.csv
    ERR: Failed to dial to host02: dial tcp 10.35.21.02:22: getsockopt: connection refused
    host01;ii  nginx                               1.8.0-1~trusty                        amd64        high performance web server
    host07;ii  nginx                               1.8.0-1~trusty                        amd64        high performance web server
    host03;ii  nginx                               1.10.1-1~trusty                       amd64        high performance web server
    host09;ii  nginx                               1.8.0-1~trusty                        amd64        high performance web server

All connection failures will have lines with "ERR:" in the beginning,
so you can easily drop not connectable hosts using "egrep -v '^ERR'".
