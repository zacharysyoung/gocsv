# [Stats as CSV](https://github.com/aotimme/gocsv/issues/64)

For the following input CSV,

```none
A,B
1,x
2,y
3,z
4,z
```

normal stats print:

```none
1. A
  Type: int
  Number NULL: 0
  Min: 1
  Max: 4
  Sum: 10
  Mean: 2.500000
  Median: 2.000000
  Standard Deviation: 1.290994
  Unique values: 4
  4 most frequent values:
      1: 1
      2: 1
      3: 1
      4: 1
2. B
  Type: string
  Number NULL: 0
  Unique values: 3
  Max length: 1
  3 most frequent values:
      z: 2
      x: 1
      y: 1
Number of rows: 4
```

CSV-stats should look something like:

| Col | Type   | NULL_ct | Min | Max | Sum | Mean | Median | Std_dev | Unique_ct | Max_str_len | No_1_val | No_1_ct | No_2_val | No_2_ct | No_3_val | No_3_ct | No_4_val | No_4_ct | No_5_val | No_5_ct |
| --- | ------ | ------: | --: | --: | --: | ---: | -----: | ------: | --------: | ----------: | -------- | ------: | -------- | ------: | -------- | ------: | -------- | ------: | -------- | ------: |
| A   | int    |       0 |   1 |   4 |  10 |  2.5 |    2.0 |    1.29 |         4 |             | 4        |       1 | 1        |       1 | 2        |       1 | 3        |       1 |          |         |
| B   | string |       0 |     |     |     |      |        |         |         3 |           1 | z        |       2 | x        |       1 | y        |       1 |          |         |          |         |

It cannot accommodate "Number of rows", but the dims and nrow subcommands handle that, particularly `dims -csv`.
