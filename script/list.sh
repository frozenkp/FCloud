#!/bin/sh

ls -lh $1 | awk 'BEGIN{a=0;}{
if(a!=0)printf("%s %s\n",$9,$5);
a++;
}'
