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
