package utils

import (
	"bytes"
	"strconv"
)

//AppendText ... Joins 2 strings like a StringBuffer in Java
func AppendText(str1 string, str2 string) string {
	var buf bytes.Buffer

	buf.WriteString(str1)
	buf.WriteString(str2)
	result := buf.String()

	return result
}

//AppendTextAndInt ... Joins an int to a string
func AppendTextAndInt(str1 string, num int) string {

	txt := strconv.Itoa(num)

	return AppendText(str1, txt)
}
