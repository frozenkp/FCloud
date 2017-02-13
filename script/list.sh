#!/bin/sh

ls -lh $1 | awk 'BEGIN{a=0;}{
if(a!=0){
  printf("%s",$9);
  for(b=10;b<=NF;b++)printf(" %s",$b);
  printf("*%s\n",$5);
}
a++;
}'
