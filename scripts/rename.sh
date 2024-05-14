#!/bin/sh

git restore cmd/cli/cli_test.go cmd/cli/main.go cmd/cli/main_test.go cmd/cli/view.go cmd/cli/view_test.go subcmd/clean/clean.go subcmd/convert/convert.go subcmd/convert/convert_test.go subcmd/cut/cut.go subcmd/cut/cut_test.go subcmd/filter/filter.go subcmd/filter/filter_test.go subcmd/head/head.go subcmd/head/head_test.go subcmd/inference.go subcmd/inference_test.go subcmd/rename/rename.go subcmd/rename/rename_test.go subcmd/sort/sort.go subcmd/sort/sort_test.go subcmd/subcmd.go subcmd/subcmd_test.go subcmd/tail/tail.go subcmd/testdata.go subcmd/testdata_test.go

subcmds="
convert
clean
cut
filter
head
rename
sort
tail
"
# "
# 

for SUBCMD in $subcmds; do
	echo '-- Change every subcmd (and test) to its own package --'
	gofmt -w -r "subcmd -> $SUBCMD" ./subcmd/$SUBCMD/*.go
	
	echo '-- Disappropriating sc name --'
	gofmt -w -r 'sc -> xx' ./subcmd/$SUBCMD/$SUBCMD.go
	
	echo '-- Renaming vars to subcmd.vars --'
	gofmt -w -r 'ColGroup -> subcmd.ColGroup'         ./subcmd/$SUBCMD/*.go
	gofmt -w -r 'FinalizeCols -> subcmd.FinalizeCols' ./subcmd/$SUBCMD/*.go
	gofmt -w -r 'Base0Cols -> subcmd.Base0Cols'       ./subcmd/$SUBCMD/*.go
	gofmt -w -r 'Base1Cols -> subcmd.Base1Cols'       ./subcmd/$SUBCMD/*.go
	gofmt -w -r 'errNoData -> subcmd.ErrNoData'       ./subcmd/$SUBCMD/*.go

	echo '-- Running goimports --'
	goimports -w ./subcmd/$SUBCMD/*.go
done



# gofmt -w -r 'subcmd -> sc' $files
 