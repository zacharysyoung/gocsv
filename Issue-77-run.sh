#!/bin/sh

# clean pretty spacing
csv clean -trim Issue-77-input.csv > input.csv

# get unique vals from Col2 and behead...
# to print just the vals
gocsv uniq -c=Col2 input.csv \
| gocsv select -c=Col2       \
| gocsv behead               \
> col2-vals.txt

# loop over vals,
# filter per val and drop Col2,
# output to val-named file
cat col2-vals.txt | while read VAL
do
gocsv filter -c=Col2 -eq $VAL input.csv \
| gocsv select -c=Col2 -exclude         \
> output-$VAL.csv
done

# print out result output CSVs in Go's txtar format
ls output*.csv | while read CSV
do
echo "-- $CSV --"
csv view $CSV
done
