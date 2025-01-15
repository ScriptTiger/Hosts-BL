[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

# Hosts-BL  
Simple tool to handle hosts file black lists that can remove comments, remove duplicates, compress to 9 domains per line, add IPv6 entries, as well as can convert black lists to multiple other black list formats compatible with other software.

Usage: `hosts-bl [options...] <source> <destination>`

Argument                  | Description
--------------------------|-----------------------------------------------------------------------------------------------------
 `-comments`              | Don't remove comments
 `-compression <number>`  | Number of domains per line, 1 to 9
 `-hash <number>`         | Hash size in bits (64\|128\|192\|256)
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

By default, dragging and dropping a hosts file over a hosts-bl executable will automatically detect the black hole address being used, pick out lines beginning with only that black hole address, assign a 64-bit hash to each host, check for and remove any duplicates, and compress it to 9 domains per line in standard hosts file format with `0.0.0.0` as the new black hole address.

In addition to removing duplicates by default, formats which support wild cards will automatically have wild card entries added and all child subdomains will be pruned from the list.

For extremely large hosts files, it may be desirable to increase the hash size to avoid collisions. The more hosts there are in a list, the more risk there is for collisions. However, at the same time, the larger the hash size is, the slower the process will run. The default 64-bit hash size is perfectly fine for reasonably-sized hosts files, such as those from Steven Black.

For additional performance gains when processing extremely large hosts files, you may want to skip checking for and removing duplicates by using the `-dupe` argument if you know the lists you are processing don't have any duplicates, such as if they are curated lists and duplicates were already previously removed. Also, using the `-from_blackhole` argument to specify the black hole address, rather than adding overhead to automatically detect it, may also slightly increase performance for larger files, as well. However, again, if you are using the Steven Black hosts files, you most likely won't notice any difference by changing these since they are reasonably sized to start with and the defaults should work perfectly fine.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/
