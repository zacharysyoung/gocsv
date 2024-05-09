# Developing

## Project Layout

Most general CSV-transforming subcommands live in pkg/subcmd. The view subcmd lives in cmd/cli because it's specific to the command-line interface.

## Testing

Most of the pkg/subcmd subcommands have their own unit tests in a complementary \_test.go file for specific functions that can easily be tested, separate from the running of the subcommand.

### Txtar tests

For aspects of a subcommand that can only be tested by calling its Run() method, the testdata directory contains a [golang.org/x/tools/txtar](https://pkg.go.dev/golang.org/x/tools/txtar) test file for each subcommand. The TestCmd func in subcmd_test.go compares txtar test file it finds in testdata with associated names and subcommands. The test files define:

- an "in" input CSV file
- a JSON file for configuring the subcommand (or the empty JSON, {}, for subcommand defaults); this JSON file has the name of the test that will print if the test fails
- a "want" file for the expected output CSV of a good run; or an "err" file with an error message to assert that the subcommand relays errors back to the caller of Run()

### Input CSV

An input file can be specified once and reused by any number of JSON-want/err pairs. If the input is simple and the developer can keep it in their head as they read the tests, then just one input file might suffice. It might be helpful to replicate the input throughout the tests to reduce scrolling/confusion. The CSV input should be simple and representative of the data, and maybe its vagueries, that the subcommand needs to work on:

- Try to limit the number of columns and rows to the proof at hand.

- Try to use a CSV consistent with these models:

  <table>
  <tr valign="top">
  <td>
  <pre>
  A
  ...
  x
  y
  z
  </pre>
  Avoid a,b,c as they mirror the column names.
  </td>

  <td>
  <pre>
  A
  1
  2
  3
  ...
  </pre>
  </td>

  <td>
  <pre>
  A
  true
  false
  </pre>
  </td>

  <td>
  <pre>
  A
  2000/1/1
  2000/1/2
  2000/1/3
  </pre>
  Favor single digits over zero-padded.
  </td>
  <td>
  <pre>
  A,B,C
  x,1,m
  y,2,n
  z,3,o
  </td>
  </tr>
  </table>

  Hopefully these will be easy to read, type, and remember/conceptualize.

  Exceptions may be unavoidable.

## Subcommand template

Subcommand's need:

- their own file in pkg/subcmd:

  ```go
  package subcmd

  // SUBCMD ... input CSV ...
  type SUBCMD struct {

  }

  func NewSUBCMD(n int, fromTop bool) *SUBCMD {
  	return &SUBCMD{

  	}
  }

  func (sc *SUBCMD) fromJSON(p []byte) error {
  	*sc = SUBCMD{}
  	return json.Unmarshal(p, sc)
  }

  func (sc *SUBCMD) CheckConfig() error {
  	return nil
  }

  func (sc *SUBCMD) Run(r io.Reader, w io.Writer) error {
  	return nil
  }
  ```

- if the subcommand has a txtar file in pkg/subcmd/testdata, it needs an entry in the subcommands map in pkg/subcmd/subcmd_test.go:

  ```go
  var subcommands = map[string]testSubCommander{
  ...
  "SUBCMD_NAME": &SUBCMD{},
  }
  ```

### Adding to cmd/cli

Adding a subcmd to the cli requires a newSUBCMD func:

```go
func newSUBCMD(args ...string) (subcmd.SubCommander, []string, error) {
	var (
		fs = flag.NewFlagSet("SUBCMD_NAME", flag.ExitOnError)
	)
	fs.Parse(args)
	return subcmd.NewSUBCMD(...), fs.Args(), nil
}
```

and entering that newSUBCMD func in a map, like streamers, that pairs the name from the command line with that func:

```go
var streamers = map[string]scMaker{
	"SUBCMD_NAME": newSUBCMD,
}
```
