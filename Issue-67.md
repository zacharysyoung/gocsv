# [replace template](https://github.com/aotimme/gocsv/issues/67)

Suggest that `replace` have a `-t` template argument which can take the matched string and capture groups in `-regex` and replaces the match to the entire regex with the value of the template. The user could specify `-repl` or `-t` but not both.

Example. If the input column `x` contains digits at the end such as `xyz23` where both `xyz` and `23` can vary from line to line then this reformats those digits with leading 0's expanding the numeric part to 5 digits, `xyz023`. Without the above feature this would require multiple lines of code picking the input apart, transforming the digits and then putting it back together. In this code `${0}` is the match. If there were capture groups then `${1}` would be the first and so on. With this feature it can be done in one line:

```
gocsv replace -c x -regex "\\d+$" -t "{{ printf \"%03d\" (atoi ${0} ) }}" myfile.csv
```

## My analysis

OP wants to transform some text in a column, padding only digits with some number of leading zeroes:

```none
Col_1      → Col_1
ab1 cd9876   ab1 cd9876
vw45 xy23    vw45 xy023
z            z
```

That cannot be accomplished with an re2 regular expression/replacement, so OP proposed using a template to process regexp matches:

```sh
gocsv replace -regex='\d+$' -template='{{ printf "%03s" $0 }}' Issue-67-input.csv
```

_The atoi func OP specified isn't necessary as printf can pad strings, but reminds me of the advantage of having access to Sprig functions inside the template._

An awk program paired with goawk could work. I don't know awk, but StackOverflow, ChatGPT, and some fiddling got me this working program, [pad.awk](./pad.awk):

```awk
BEGIN {
  padWidth = 3
  padZeroes = "000000000000"
  if (padWidth > length(padZeroes)) {
    print "error: program params padWidth=" padWidth " greater than length of padZeroes=\"" padZeroes "\""
    exit 1
  }
}
{
  # copy field to prefix
  prefix = $1
  # remove the trailing digits from prefix
  sub(/[0-9]+$/, "", prefix)
  # check if prefix differs from the field (did field actually have digits)
  if (prefix != $1) {
    # bypass prefix to get digits from field
    num = substr($1, length(prefix) + 1)
    # pad
    pad = substr(padZeroes, 0, padWidth-length(num))
    # replace field with concatenation of prefix+pad+digits
    $1 = prefix pad num
  }
  print
}
```

```sh
goawk --csv -f pad.awk Issue-67-input.csv
```

```none
Col_001
ab1 cd9876
vw45 xy023
z
```

That script has some problems:

- mistakenly modifies the header
- wouldn't work if we wanted to affect multiple matches inside a single field (e.g., remove the EOL anchor (`$`) in `[0-9]+$` to get `ab001 cd9876` and `vw045 xy023`)

I know a better awk program could do it, but a template _scheme_ could make this trivial:

```sh
gocsv replace -regex='\d+$' -template='{{ printf "%03s" $0 }}' Issue-67-input.csv
```

```none
Col_1
ab1 cd9876
vw45 xy023
z
```

Affecting all digits:

```sh
gocsv replace -regex='\d+' -template='{{ printf "%03s" $0 }}' Issue-67-input.csv
```

```none
Col_1
ab001 cd9876
vw045 xy023
z
```

My main concern is that we have to write our own replacement logic to account for a template being applied to multiple matches, i.e., our own small, imperative program to achieve the _intent_ of the declaration. And while OP's case does make sense, I cannot say for certain this intent is general enough to pick a specific implementation.
