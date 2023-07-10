package fountain

import (
	"encoding/json"
	"errors"
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

func serializeCodeBlock(codeBlocks []LTBlock) []byte{
	bytes, err := json.Marshal(codeBlocks)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func deserializeCodeBlock(bytes []byte) ([]LTBlock){
	var newLTBlock []LTBlock
	err := json.Unmarshal(bytes, &newLTBlock)
	if err != nil {
		log.Fatal(err)
	}

	return newLTBlock
}

func encode(message []byte, kSize int, encodedSymbolSize int, seed int64) []byte{
	c := NewRaptorCodec(kSize, 4)
	ids := make([]int64, encodedSymbolSize)
	random := rand.New(rand.NewSource(seed))

	messageCopy := make([]byte, len(message))
	copy(messageCopy, message)

	for i := range ids {
		ids[i] = int64(random.Intn(60000))
	}

	encodedBlocks := EncodeLTBlocks(messageCopy, ids, c)
	serializedEncodedBlocks := serializeCodeBlock(encodedBlocks)
	
	return serializedEncodedBlocks
}

func decode(serializedEncodedBlocks [][]byte, kSize int, messageSize int) ([]byte, error){
	c := NewRaptorCodec(kSize, 4)
	decoder := newRaptorDecoder(c.(*raptorCodec), messageSize)

	for i := 0; i < len(serializedEncodedBlocks); i++ {
		EncodedBlocks := deserializeCodeBlock(serializedEncodedBlocks[i])
		for j := 0; j < len(EncodedBlocks); j++ {
			if (decoder.AddBlocks([]LTBlock{EncodedBlocks[j]})) {
				out := decoder.Decode()
				return out, nil
			}
		}
	}

	return nil, errors.New("Not Decoding!")
}

func TestAvailableKsize(t *testing.T) {
	availableKsize := make([]int, 0)
	filePath := "dummy_data/4c3_data/dummy1.json"
	message, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	for kSize := 4; kSize < 100; kSize++ {
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

func TestMessageED(t *testing.T) {
	filePath := "dummy_data/50MB_dummy.json"
	message, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	kSize := 45
	encodedSymbolSize := 20

	serializedEncodedBlocks := make([][]byte, 3)

	
	serializedEncodedBlocks[0] = encode(message, kSize, encodedSymbolSize, 8923483)
	fmt.Println(len(serializedEncodedBlocks[0]))
	serializedEncodedBlocks[1] = encode(message, kSize, encodedSymbolSize, 8923486)
	serializedEncodedBlocks[2] = encode(message, kSize, encodedSymbolSize, 8923487)

	out, err := decode(serializedEncodedBlocks, kSize, len(message))

	if err != nil {
		t.Errorf(err.Error())
	}
	
	if !reflect.DeepEqual(message, out) {
		t.Errorf("Different!")
	}else{
		var jsonData []Data
		err = json.Unmarshal(out, &jsonData)

		if err != nil {
			t.Errorf(err.Error())
		}
		
		outputJsonFile(jsonData)
	}
}

func Test4c3Combination(t *testing.T){
	json1Path := "dummy_data/4c3_data/dummy1.json"
	json2Path := "dummy_data/4c3_data/dummy2.json"
	kSize := 7
	encodedSymbolSize := 3


	message1, err := ioutil.ReadFile(json1Path)
	if err != nil {
		log.Fatal(err)
	}

	message2, err := ioutil.ReadFile(json2Path)
	if err != nil {
		log.Fatal(err)
	}
	
	producer1 := encode(message1, kSize, encodedSymbolSize, 8923483)
	producer2 := encode(message1, kSize, encodedSymbolSize, 8923484)
	producer3 := encode(message1, kSize, encodedSymbolSize, 9234855)
	producer4 := encode(message2, kSize, encodedSymbolSize, 8923486) // malicious

	serializedEncodedBlocks1 := make([][]byte, 3)
	serializedEncodedBlocks1[0] = producer2
	serializedEncodedBlocks1[1] = producer3
	serializedEncodedBlocks1[2] = producer4

	serializedEncodedBlocks2 := make([][]byte, 3)
	serializedEncodedBlocks2[0] = producer1
	serializedEncodedBlocks2[1] = producer3
	serializedEncodedBlocks2[2] = producer4

	serializedEncodedBlocks3 := make([][]byte, 3)
	serializedEncodedBlocks3[0] = producer1
	serializedEncodedBlocks3[1] = producer2
	serializedEncodedBlocks3[2] = producer4

	serializedEncodedBlocks4 := make([][]byte, 3) // Not malicious
	serializedEncodedBlocks4[0] = producer1
	serializedEncodedBlocks4[1] = producer2
	serializedEncodedBlocks4[2] = producer3

	out1, err := decode(serializedEncodedBlocks1, kSize, len(message1))
	out2, err := decode(serializedEncodedBlocks2, kSize, len(message1))
	out3, err := decode(serializedEncodedBlocks3, kSize, len(message1))
	out4, err := decode(serializedEncodedBlocks4, kSize, len(message1))

	if err != nil {
		t.Errorf(err.Error())
	}

	var jsonData []Data
	err = json.Unmarshal(out1, &jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("out1: %v\n", string(out1))

	err = json.Unmarshal(out2, &jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("out2: %v\n", string(out2))

	err = json.Unmarshal(out3, &jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("out3: %v\n", string(out3))

	err = json.Unmarshal(out4, &jsonData)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("out4: %v\n", string(out4))

}