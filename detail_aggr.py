#!/usr/bin/python

import sys

stat = dict()
total = 0

for line in sys.stdin:
    total += 1
    fields = line.split(' ')
    if len(fields) != 3:
        raise ValueError("incorrect line %d: %s" % (total, line))

    s = stat.get(fields[0])
    if s is None:
        s = 1
    else:
        s += 1

    stat[fields[0]] = s

for k in sorted(stat.keys()):
    sys.stdout.write("%s %d\n" % (k, stat[k]))

sys.stdout.write("total %d\n" % total)
