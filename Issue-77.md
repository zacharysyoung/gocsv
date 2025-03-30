# [splitting based on a column](https://github.com/aotimme/gocsv/issues/77)

Would be nice if one could split an input file into separate output files based on a grouping column (as opposed being based on the number of records). It would be good enough if the rows for each output file must be contiguous in the input. The value of the grouping column could be a suffix of the base of the generated filenames.

An option to omit the grouping column in the output files would also be nice.

For example, consider the situation where one had a file of daily data, one row per day, with a column for the year. There would be many rows for each year and what we want is to split the file into separate files, one for each year as opposed to splitting it based on the max number of records in each output file. The generated filenames would be something like myfile-2024.csv, myfile-2025.csv assuming that the input file is myfile.csv and it has the indicated two years in the `year` column of the input.

In miller one can do it like this, assuming a grouping column named `year`. There is no option to omit the grouping column from the output files.

```
mlr --csv --from myfile.csv split -g year
```

An inverse function would be nice too in case one split by a grouping column but the grouping column had been omitted from the resulting files. It would take the resulting files and restore the file that the split files were derived from getting the grouping column from the filenames in the case that the grouping column had been omitted by the split. This inverse function would be lower priority.

## [Comment](https://github.com/aotimme/gocsv/issues/77#issuecomment-2764762726)

In reading about the `stack` subcommand I think that already would be the inverse so that part is already done.

## My analysis

Given an input like:

```none
-- input.csv --
ID, Col2, Col3
 1, a,    blue
 2, b,    blue
 3, a,    red
 4, c,    green
 5, c,    yellow
```

if OP specified "split on Col2", they want 3 named output files (based on the three unique values a, b, c), each file with just the rows that match their respective unique values, and Col2 dropped (since the name tells OP the value):

```none
-- output-a.csv --
ID, Col3
 1, blue
 3, red
-- output-b.csv --
ID, Col3
 2, blue
-- output-c.csv --
ID, Col3
 4, green
 5, yellow
```

For, the inverse, OP could use the stack subcommand to recombine them:

```sh
gocsv stack -group-name Col2 -groups a,b,c output* \
| gocsv select -c=ID,Col2,Col3                     \
| gocsv sort -c=ID                                 \
> stacked.csv
```

```none
-- stacked.csv --
ID, Col2, Col3
 1, a,    blue
 2, b,    blue
 3, a,    red
 4, c,    green
 5, c,    yellow
```

## My response

I don't think GoCSV should do this.

I understand that given an input like the following (ignore the pretty-fied spaces):

```none
-- input.csv --
ID, Col2, Col3
 1, a,    blue
 2, b,    blue
 3, a,    red
 4, c,    green
 5, c,    yellow
```

you want to be able to run something like `gocsv split -c=Col2 input.csv`, and get the following three files:

```none
-- output-a.csv --
ID, Col3
 1, blue
 3, red
-- output-b.csv --
ID, Col3
 2, blue
-- output-c.csv --
ID, Col3
 4, green
 5, yellow
```

I believe GoCSV wants to be a general purpose toolkit for processing CSVs. While GoCSV never states that it follows the [Unix Philosophy](https://en.wikipedia.org/wiki/Unix_philosophy), that was always the vibe I got:

> The tool is built for pipelining, so most commands accept a CSV from standard input and output to standard output.

I understand the desire to have GoCSV do more, but this seems overly specific/specialized to me, and the decision to drop the column(s) seems like a very personal one that could go either way depending on any number of factors.

GoCSV can already accomplish what you've outlined with a series of commands/pipelines:

```sh
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
    gocsv filter -c=Col2 -eq "$VAL" input.csv \
    | gocsv select -c=Col2 -exclude           \
    > "output-$VAL.csv"
done
```

What do you think, @alden?
