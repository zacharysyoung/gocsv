# CmdDev2

Testing out new ideas for how to develop and test a package, and a command-line interface around a package.

- The GoCSV-like subcommands live in [./pkg/subcmd](./pkg/subcmd).
- The GoCSV-like command itself lives in [./cmd/cli](./cmd/cli) and uses subcommands that satisfy the Subcommander interface. The frontend-specific view subcommand also lives there.

The cli manages all command-line aspects like getting the ins and outs, defining and parsing flags and args, and halting with errors. The individual subcommands are mostly only responsible for running themselves against an io.Reader and io.Writer, and bubbling errors up.

## Testing

Both the cli and subcmds make use of golang.org/x/tools/txtar to validate ins and outs, and also validate behavior like handling/printing errors. The testdata folders in both cli and subcmd have a list of txtar .txt files. Each .txt file holds tests for a specific subcommand; inside each .txt file:

- at least one '-- in --' file is specified that represents the input data to test
- subsequent pairs of files can reference the prior '-- in --' file
  - the first file in the pair denotes some kind of configuration: for cli this will be command-line args; for subcmd it will be JSON that each subcmd can unmarshal into itself

Each .txt file name becomes the name of a subtest, and each txtar configuration file name becomes the name of subtest under that.

For subcmd/testdata/filter.txt, the following will run the two tests filter/eq and filter/less-than against the input in '-- in --', using the JSON configuration for each test:

```none
-- in --
H
A
a
-- eq --
{"Col": 1, "Operator": "eq", "Value": "a"}
-- want --
H
a
-- less-than --
{"Col": 1, "Operator": "lt", "Value": "a"}
-- want --
H
A
```

For cmd/cli/testdata/view.txt, the following will run the test view/markdown, interpreting the second file as command-line args:

```none
-- want --
Col1, Col2,    Col3$
   1, longish, a   $
   2, short,   bb  $
 3.0, foo,     ccc $
-- as markdown --
-md
-- want --
| Col1 | Col2    | Col3 |
| ---: | ------- | ---- |
|    1 | longish | a    |
|    2 | short   | bb   |
|  3.0 | foo     | ccc  |
```

The in-file also makes use of terminal "$" markers to denote the end of padded fields (and also prevent autoformatters from trimming trailing whitespace).  Both test harnesses preprocess all inputs and remove terminal "$"s.

For pkg/subcmd, the test harness also allows for aligned input and output text files, for easier viewing of columns. It handles the de-formatting of the input and formatting of the output:

```none
-- in --
A, B, C, D
1, 2, 3, 4
-- exclude --
{"Cols":[2, 3], "Exclude":true}
-- want --
A, D
1, 4
```

The input text gets transformed to:

```none
A,B,C,D
1,2,3,4
```

before being run through the select subcommand and compared to the expected:

```none
A,D
1,4
```

Subtle whitespace issues can go undetected in the aligned views. The pkg/subcmd test harness also defines a -quote flag that can be passed to 'go test' which falls back on Sprintf("%q", output) to succinctly and unambiguously show both got and want.
