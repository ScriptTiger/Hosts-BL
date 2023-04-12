[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

# Hosts-BL  
Simple tool to handle hosts file black lists that can remove comments, remove duplicates, compress to 9 domains per line, add IPv6 entries, as well as can convert black lists to multiple other black list formats compatible with other software.

Usage: `hosts-bl [options...] <source> <destination>`

Argument                  | Description
--------------------------|-----------------------------------------------------------------------------------------------------
 `-comments`              | Don't remove comments
 `-compression <number>`  | Number of domains per line, 1 to 9
 `-dupe`                  | Don't check for and remove duplicates
 `-f <format>`            | Destination format:
--------------------------| **adblock, dnsmasq, dualserver, fqdn,**
--------------------------| **hosts, ipv6, privoxy, rfqdn, rpz, unbound**
 `-from_blackhole <IPv4>` | Black hole address in source
 `-i <file>`              | Source file
 `-o <file>`              | Destination file
 `-to_blackhole <IPv4>`   | Black hole address in destination
 `-to_blackhole_v6 <IPv6>`| IPv6 Black hole address in destination

`-` can be used in place of `<file>` to designate standard input as the source and/or standard output as the destination. If standard input is used, standard output will be used by default if no destination file is given.

By default, dragging and dropping a hosts file over a hosts-bl executable will automatically pick out lines beginning with `0.0.0.0`, check for and remove any duplicates, and compress it to 9 domains per line in standard hosts file format with `0.0.0.0` as the black hole address.

In addition to removing duplicates by default, formats which support wild cards will automatically have wild card entries added and all child subdomains will be pruned from the list.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=MZ4FH4G5XHGZ4)

Donate Monero (XMR): 441LBeQpcSbC1kgangHYkW8Tzo8cunWvtVK4M6QYMcAjdkMmfwe8XzDJr1c4kbLLn3NuZKxzpLTVsgFd7Jh28qipR5rXAjx
