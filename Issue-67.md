# Replace w/templates

OP wants to transform some text in a column, padding only digits with some number of leading zeroes:

```none
Col_1 → Col_1
ab1     ab001
xy23    xy023
z       z
```

That cannot be accomplished with an re2 regular expression/replacement, so OP proposed using a template to process regexp matches:

- **-regex**: `\d+$`
- **-templ**: `{{ printf "%03d" (atoi $0) }}`

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

- **-regex**: `\d+`
- **-templ**: `{{ printf "%03s" $0 }}`

```none
Col_1      → Col_1
ab1 cd987    ab001 cd987
vw56 xy23    vw056 xy023
z            z
```

My main concern is that we have to write our own replacement logic to account for a template being applied to multiple matches.
