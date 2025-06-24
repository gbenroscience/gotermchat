package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/oklog/ulid"
)

// User ... Models information for the User
type User struct {
	ID       string    `bson:"_id" json:"id"`
	Name     string    `json:"name"`
	RegTime  time.Time `json:"regTime"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
}

type LoginResponse struct {
	User    User   `json:"user"`
	Message string `json:"msg"`
}

type GroupMake struct {
	Name string `json:"name"`
	//Will double as the group id
	Alias string `json:"alias"`
}
type GroupAdd struct {
	NameOrAlias string `json:"name_or_alias"`
	Phone       string `json:"phone"`
}
type GroupMessage struct {
	NameOrAlias string `json:"name_or_alias"`
	TextMessage string `json:"msg"`
}
type GroupRemoveMember struct {
	NameOrAlias string `json:"name_or_alias"`
	MemberPhone string `json:"member_phone"`
}
type GroupDelete struct {
	NameOrAlias string `json:"name_or_alias"`
}
type GroupsListForUser struct {
	Phone string `json:"phone"`
}

// AppendText ... Joins 2 strings like a StringBuffer in Java
func AppendText(str1 string, str2 string) string {
	var buf bytes.Buffer

	buf.WriteString(str1)
	buf.WriteString(str2)
	result := buf.String()

	return result
}

// AppendTextAndInt ... Joins an int to a string
func AppendTextAndInt(str1 string, num int) string {

	txt := strconv.Itoa(num)

	return AppendText(str1, txt)
}

func GenUlid() string {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	return id.String()
}

func EncodeStruct(b interface{}) (string, error) {
	s, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		return "", err
	}
	return string(s), err
}

func DumpStruct(b interface{}) error {
	s, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		log.Printf("Error dumping struct: %v\n", err)
		return err
	}
	log.Println("Dumping struct: \n", string(s))
	return err
}

// DecodeItem Decodes a json string into a pointer to a generic Golang struct. Pass a pointer to this function
func DecodeItem(jsn string, destPtr interface{}) error {
	return json.NewDecoder(bytes.NewBufferString(jsn)).Decode(destPtr)
}

// DecodeBytes Decodes json bytes into a pointer to a generic Golang struct. Pass a pointer to this function
func DecodeBytes(jsn []byte, destPtr interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(jsn)).Decode(destPtr)
}
