# [how to replace character with nonprinting character](https://github.com/aotimme/gocsv/issues/78)

Suppose we have this file:

--- newline.csv ---

```
abc
"one
two"
```

and we want to replace newlines with hex ef (or other non-printing character). How can we do this? This works to replace newline with ! but I haven't figured out how to express non-printing characters with the -repl flag.

```
gocsv replace -regex \n -repl ! newline.csv
```

giving this as expected

```
a
one!two
```

I have tried using \xef, \u00ef and variations of those but haven't found where this is documented or how to do it.

(Above we used Windows cmd so quoting and escaping may be different on other platforms.)

## My analysis

OP wants to do something like replace all lower-case z chars with an upper-case Z (\u005A):

```none
-- Issue-78-input.csv --
1
baz
```

```none
gocsv replace -regex='z' -repl='\u005A' Issue-78-input.csv
```

and get:

```none
1
baZ
```

This won't work right now because "\u005A" is string of six literal chars: "\", "u" ... "A".

In order for the replace to command to interpret those six chars as a single unicode code point, it would need to unquote the string, like we do for interpreting delimiters:

```go
repl, err := strconv.Unquote(`"` + sub.repl + `"`)
```

## My response

Good question. You were right to try an escaped code point, like "\u00ef" , but inside the replace subcommand the runtime sees 6 literal characters: "\", "u", "0", "0", "e", "f", and not the single code point you intended.

I'll mark this as a bug, but I cannot guarantee a fix... semantics of unquoting/reinterpreting the string... but I think we can fix it.

In the meantime, you could try to get your shell to put the literal value in a variable, and then pass that variable to the -repl flag.

In zsh, replacing the lower-case z with an upper-case Z (\u005A) looks like:

```none
-- input.csv --
1
baz
```

```zsh
#!/bin/zsh

set -e

literal=$(echo '\U005A')

gocsv replace -regex='z' -repl=$literal input.csv
```

and that gets me:

```none
1
baZ
```

I don't have access to a Windows Command prompt, but this [SuperUser answer](https://superuser.com/a/1858992/96227) to _How to put unicode in Cmd/Batch?_ might help you accomplish the same thing.
