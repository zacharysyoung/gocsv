# CmdDev

Testing out new ideas for how to develop and test a package, and a command-line interface around a package.

- The GoCSV-like subcommands live in [./pkg/subcmd](./pkg/subcmd).
- The GoCSV-like command itself lives in [./cli](./cli) and uses subcommands that satisfy the Subcommander interface.

The cli manages all command-line aspects like getting the ins and outs, defining and parsing flags and args, and halting with errors. The individual subcommands are mostly only responsible for running themselves against an io.Reader and io.Writer, and bubbling errors up.

Both the cli and subcmds make use of golang.org/x/tools/txtar to validate ins and outs, and also behavior like handling/printing errors. The testdata folders in both cli and pkg/subcmd have a list of txtar files that follow this format:

- each \_cmd.txt txtar file holds tests for a specific subcommand; inside each \_cmd.txt file:
  - at least on -- in -- file is specified that represents the input data to test
  - subsequent pairs of files can reference the prior -- in -- file. The first file in the pair denotes some kind of configuration that
