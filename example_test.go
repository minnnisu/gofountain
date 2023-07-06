package fountain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

type Data struct {
	Name        string `json:"name"`
	Account  	string `json:"account"`
}

func Encode(message []byte, e int, c Codec, seed int64) []LTBlock{
	ids := make([]int64, e)
	random := rand.New(rand.NewSource(seed))
	for i := range ids {
		ids[i] = int64(random.Intn(60000))
	}
	codeBlocks := EncodeLTBlocks(message, ids, c)

	return codeBlocks
}

func outputJsonFile(jsonData []Data){
	// JSON 파일 생성
	filePath := "output.json"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// JSON 인코딩
	encoder := json.NewEncoder(file)
	err = encoder.Encode(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	// 파일에 JSON 데이터를 쓴 후, 파일 경로 출력
	log.Printf("JSON data has been written to %s\n", filePath)
}

func compareTwoJson(json1 []Data, json2 []Data){
	for j := 0; j < len(json2); j++ {
		fmt.Println(j)
		if(json1[j].Name != json2[j].Name){
			fmt.Printf("original name: %v, recovered name: %v", json1[j].Name , json2[j].Name)
		}

		if(json1[j].Account != json2[j].Account){
			fmt.Printf("original account: %v, recovered account: %v", json1[j].Account , json2[j].Account)
		}
	}	
}

func TestAvailableKsize(t *testing.T) {
	availableKsize := make([]int, 0)
	filePath := "dummy_data/50MB_dummy.json"
	message, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	for kSize := 100; kSize < 110; kSize++ {
		c := NewRaptorCodec(kSize, 4)
		ids := make([]int64, kSize + 50)
		random := rand.New(rand.NewSource(8923489))
		for i := range ids {
			ids[i] = int64(random.Intn(60000))
		}
	
		messageCopy := make([]byte, len(message))
		copy(messageCopy, message)
	
		codeBlocks := EncodeLTBlocks(messageCopy, ids, c)
		
		t.Log("DECODE--------")
		decoder := newRaptorDecoder(c.(*raptorCodec), len(message))
		for i := 0; i < kSize+50; i++ {
			decoder.AddBlocks([]LTBlock{codeBlocks[i]})
		}
		if decoder.matrix.determined() {
			out := decoder.Decode()
			
			if !reflect.DeepEqual(message, out) {
				fmt.Printf("Not Equal: kSize %v\n", kSize)
			}else {
				availableKsize = append(availableKsize, kSize)
			}
		}else {
			fmt.Printf("Not determined: kSize %v \n", kSize)
		}
	}

	fmt.Printf("AvailableKsize: %v\n", availableKsize)
}

func serializeCodeBlockToByte(codeBlocks []LTBlock) []byte{
	bytes, err := json.Marshal(codeBlocks)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func serializeByteToCodeBlock(bytes []byte) ([]LTBlock){
	var newLTBlock []LTBlock
	err := json.Unmarshal(bytes, &newLTBlock)
	if err != nil {
		log.Fatal(err)
	}

	return newLTBlock
}

func TestMessageED(t *testing.T) {
	// JSON 파일 열기
	filePath := "dummy_data/50MB_dummy.json"
	message, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	c := NewRaptorCodec(45, 4)

	messageCopy1 := make([]byte, len(message))
	copy(messageCopy1, message)

	messageCopy2 := make([]byte, len(message))
	copy(messageCopy2, message)

	codeBlocks1 := Encode(messageCopy1, 100, c, 8923485)
	codeBlocks2 := Encode(messageCopy2, 100, c, 8923487)

	byteData1 := serializeCodeBlockToByte(codeBlocks1)
	reCodeBlock1 := serializeByteToCodeBlock(byteData1)

	byteData2 := serializeCodeBlockToByte(codeBlocks2)
	reCodeBlock2 := serializeByteToCodeBlock(byteData2)

	t.Log("DECODE--------")
	decoder := newRaptorDecoder(c.(*raptorCodec), len(message))

	i := 0
	for !decoder.matrix.determined() && i < 70 {
		fmt.Println(decoder.AddBlocks([]LTBlock{reCodeBlock1[i]}))
		fmt.Println(decoder.AddBlocks([]LTBlock{reCodeBlock2[i+1]}))
		i += 2
	}

	if decoder.matrix.determined() {
		// t.Log("Recovered:\n", decoder.matrix.v)
		out := decoder.Decode()

		// // JSON 데이터를 Go 데이터 구조로 언마샬링
		// var recoveredJsonData []Data
		// err = json.Unmarshal(out, &recoveredJsonData)

		// if err != nil {
		// 	log.Printf("error decoding sakura response: %v", err)
		// 	if e, ok := err.(*json.SyntaxError); ok {
		// 		log.Printf("syntax error at byte offset %d", e.Offset)
		// 	}
		// }

		if !reflect.DeepEqual(message, out) {
			t.Errorf("Not different")
		}
	}else {
		t.Errorf("Not determined")
	}
}