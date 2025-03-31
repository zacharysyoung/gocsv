# [replace template](https://github.com/aotimme/gocsv/issues/67)

Suggest that `replace` have a `-t` template argument which can take the matched string and capture groups in `-regex` and replaces the match to the entire regex with the value of the template. The user could specify `-repl` or `-t` but not both.

Example. If the input column `x` contains digits at the end such as `xyz23` where both `xyz` and `23` can vary from line to line then this reformats those digits with leading 0's expanding the numeric part to 5 digits, `xyz00023`. Without the above feature this would require multiple lines of code picking the input apart, transforming the digits and then putting it back together. In this code `${0}` is the match. If there were capture groups then `${1}` would be the first and so on. With this feature it can be done in one line:

```
gocsv replace -c x -regex "\\d+$" -t "{{ printf \"%05d\" (atoi ${0} ) }}" myfile.csv
```

## My analysis

OP wants to transform some text in a column, padding only digits with some number of leading zeroes:

```none
Col_1 → Col_1
ab1     ab001
xy23    xy023
z       z
```

That cannot be accomplished with an re2 regular expression/replacement, so OP proposed using a template to process regexp matches:

```sh
gocsv replace -regex='\d+$' -templ='{{ printf "%03d" (atoi $0) }}'
```

The atoi func isn't necessary as printf can pad strings, `{{ printf "%03s" $0 }}`, but reminds me of the advantage of having access to Sprig functions inside the template.

An awk program paired with goawk could work. I don't know awk, but StackOverflow and ChatGPT got me this working program:

```awk
prefix = $1
sub(/[0-9]+$/, "", prefix)
if (length(prefix) != length($1)) {
  num = substr($1, length(prefix) + 1)
  pad = substr("00000", 0, 5-length(num))
  $1 = prefix pad num
}
print
```

That wouldn't work if I wanted to affect multiple matches inside a single field (by removing the end-of-string anchor). I'm sure awk could do it, but a template _scheme_ could make this trivial:

```sh
gocsv replace -regex='\d+' -templ='{{ printf "%03s" $0 }}'
```

```none
Col_1      → Col_1
ab1 cd987    ab001 cd987
vw56 xy23    vw056 xy023
z            z
```

My main concern is that we have to write our own replacement logic to account for a template being applied to multiple matches.
