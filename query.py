#!/usr/bin/python

import sys

f1_name = sys.argv[1]
f2_name = sys.argv[2]

f1_stat = dict()
f1_count = 0

f2_stat = dict()
f2_count = 0

stat = dict()

new_count = 0
miss_count = 0

change_count = 0

f1 = open(f1_name, "r")
f2 = open(f2_name, "r")
 
for line in f1:
    line = line.rstrip()
    f1_count += 1
    s = f1_stat.get(line, 0)
    s += 1
    f1_stat[line] = s

f1.close()

for line in f2:
    line = line.rstrip()
    f2_count += 1
    s = f2_stat.get(line, 0)
    s += 1
    f2_stat[line] = s

f2.close()

for k in sorted(f1_stat.keys()):
    f1_s = f1_stat[k]
    f2_s = f2_stat.get(k)
    if f2_s is None:
      stat[k] = "-   "
      miss_count += 1
    elif f1_s != f2_s:
      stat[k] = " %d->%d" % (f1_s, f2_s)
      change_count += 1

for k in sorted(f2_stat.keys()):
    f1_s = f1_stat.get(k)
    if f1_s is None:
      stat[k] = "+   "
      new_count += 1

for k in sorted(stat.keys()):
    sys.stdout.write("%s %s\n" % (stat[k], k))

sys.stdout.write("%s total %d\n" % (f1_name, f1_count))
sys.stdout.write("%s total %d\n" % (f2_name, f2_count))
sys.stdout.write("%s count mismatch  %d\n" % (f1_name, change_count))
sys.stdout.write("%s miss            %d\n" % (f1_name, miss_count))
sys.stdout.write("%s new             %d\n" % (f1_name, new_count))
