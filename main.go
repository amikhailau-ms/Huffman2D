package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
)

const (
	rusalphabet  = "абвгдежзийклмнопрстуфхцчшщьыэюя "
	rusSpaceStat = 0.175
	inFileName   = "in2.txt"
	outFileName  = "out.txt"
)

var (
	rusalphabetstats = []float64{0.062, 0.014, 0.038, 0.013, 0.025, 0.072, 0.007, 0.016,
		0.062, 0.010, 0.028, 0.035, 0.026, 0.053, 0.090, 0.023, 0.040, 0.045, 0.053, 0.021, 0.002,
		0.009, 0.004, 0.012, 0.006, 0.003, 0.014, 0.016, 0.003, 0.006, 0.018, 0.175}
	totalProb = 0.0
)

type Node struct {
	leftson     *Node
	rightson    *Node
	symbol      rune
	probability float64
}

type CodingTable struct {
	table      map[rune]string
	avgWordLen float64
	entropy    float64
}

type NodeString struct {
	leftson     *NodeString
	rightson    *NodeString
	symbol      string
	probability float64
}

type CodingTableString struct {
	table      map[string]string
	avgWordLen float64
	entropy    float64
}

func main() {

	pwd, _ := os.Getwd()
	textByte, err := ioutil.ReadFile(pwd + "/" + inFileName)
	if err != nil {
		fmt.Println("Не удалось открыть входной файл.")
		os.Exit(1)
	}
	text := strings.ToLower(string(textByte))
	rusalphabetRune := []rune(rusalphabet)

	realProbs := make([]float64, 32, 32)
	polishedText := ""
	for _, symbol := range text {
		if index := findRuneInSlice(rusalphabetRune, symbol); index != -1 || symbol == 'ъ' || symbol == 'ё' {
			if symbol == 'ё' {
				polishedText += "e"
				realProbs[5] += 1.0
			} else if symbol == 'ъ' {
				polishedText += "ь"
				realProbs[26] += 1.0
			} else {
				polishedText += string(symbol)
				realProbs[index] += 1.0
			}
		}
	}
	textLength := len([]rune(polishedText))
	for i := range realProbs {
		realProbs[i] /= float64(textLength)
	}

	reportFile, err := os.Create(pwd + "/" + outFileName)
	if err != nil {
		fmt.Println("Не удалось открыть выходной файл.")
		os.Exit(1)
	}
	defer reportFile.Close()
	reportWriter := bufio.NewWriter(reportFile)
	defer reportWriter.Flush()
	reportWriter.WriteString("Результат: \n\n")

	nodes := convertStatsToNodes(rusalphabetstats)
	root := buildHuffmanTree(nodes)
	thCodingTable := makeCodingTable(root, realProbs)
	reportWriter.WriteString(fmt.Sprintf("Табличные вероятности:\nЭнтропия: %.3f.\nСредняя длина слова: %.3f.\n\n",
		thCodingTable.entropy, thCodingTable.avgWordLen))

	realNodes := convertStatsToNodes(realProbs)
	root = buildHuffmanTree(realNodes)
	prCodingTable := makeCodingTable(root, realProbs)
	reportWriter.WriteString(fmt.Sprintf("Вероятности по тексту:\nЭнтропия: %.3f.\nСредняя длина слова: %.3f.\n\n",
		prCodingTable.entropy, prCodingTable.avgWordLen))

	if prCodingTable.avgWordLen > thCodingTable.avgWordLen {
		reportWriter.WriteString(fmt.Sprintf("Использование табличных вероятностей оказалось лучше.\n\n"))
	} else if prCodingTable.avgWordLen < thCodingTable.avgWordLen {
		reportWriter.WriteString(fmt.Sprintf("Определение вероятностей по тексту оказалось лучше.\n\n"))
	} else {
		reportWriter.WriteString(fmt.Sprintf("Оба метода дали одинаковый результат.\n\n"))
	}

	pairs := buildPairs()
	countPairProbabilities(pairs, polishedText)
	pairNodes := convertPairStatsToNodes(pairs)
	rootPair := buildHuffmanPairTree(pairNodes)
	pairCodingTable := makePairCodingTable(rootPair)
	reportWriter.WriteString(fmt.Sprintf("Двумерный код Хаффмана:\nЭнтропия: %.3f.\nСредняя длина слова: %.3f.\n\n",
		pairCodingTable.entropy, pairCodingTable.avgWordLen))

	reportWriter.WriteString(fmt.Sprintf("При сравнении с одномерным кодом нужно учесть, что каждое слово в двумерном коде шифрует сразу 2 символа.\n"))

	if prCodingTable.avgWordLen > pairCodingTable.avgWordLen/2 {
		reportWriter.WriteString(fmt.Sprintf("Использование двумерного кода оказалось лучше.\n\n"))
	} else if prCodingTable.avgWordLen < pairCodingTable.avgWordLen/2 {
		reportWriter.WriteString(fmt.Sprintf("Использование одномерного кода оказалось лучше.\n\n"))
	} else {
		reportWriter.WriteString(fmt.Sprintf("Оба метода дали одинаковый результат.\n\n"))
	}
}

func convertStatsToNodes(letterStats []float64) []*Node {
	result := make([]*Node, 0, 70)
	rusalphabetRune := []rune(rusalphabet)
	for _, s := range rusalphabetRune {
		result = append(result, &Node{symbol: s, probability: letterStats[findRuneInSlice(rusalphabetRune, s)]})
	}
	return result
}

func buildHuffmanTree(nodes []*Node) *Node {
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].probability < nodes[j].probability })
	for len(nodes) > 1 {
		firstNode := nodes[0]
		secondNode := nodes[1]
		nodes = nodes[2:]
		newNode := &Node{
			symbol:      '-',
			probability: firstNode.probability + secondNode.probability,
			leftson:     firstNode,
			rightson:    secondNode,
		}
		nodes = append(nodes, newNode)
		sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].probability < nodes[j].probability })
	}
	return nodes[0]
}

func buildHuffmanPairTree(nodes []*NodeString) *NodeString {
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].probability < nodes[j].probability })
	for len(nodes) > 1 {
		firstNode := nodes[0]
		secondNode := nodes[1]
		nodes = nodes[2:]
		newNode := &NodeString{
			symbol:      "-",
			probability: firstNode.probability + secondNode.probability,
			leftson:     firstNode,
			rightson:    secondNode,
		}
		nodes = append(nodes, newNode)
		sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].probability < nodes[j].probability })
	}
	return nodes[0]
}

func makeCodingTable(root *Node, realProbs []float64) *CodingTable {
	totalProb = 0.0
	codingTable := &CodingTable{
		avgWordLen: 0.0,
		entropy:    0.0,
		table:      make(map[rune]string),
	}
	dfs(root, codingTable, "", realProbs)
	if totalProb < 0.99 || totalProb > 1.01 {
		fmt.Printf("WRONG PROBS!!! %v", totalProb)
		os.Exit(-3)
	}
	return codingTable
}

func makePairCodingTable(root *NodeString) *CodingTableString {
	totalProb = 0.0
	codingTable := &CodingTableString{
		avgWordLen: 0.0,
		entropy:    0.0,
		table:      make(map[string]string),
	}
	dfsPair(root, codingTable, "")
	if totalProb < 0.99 || totalProb > 1.01 {
		fmt.Printf("WRONG PROBS!!! %v", totalProb)
		os.Exit(-3)
	}
	return codingTable
}

func dfs(node *Node, codingTable *CodingTable, code string, realProbs []float64) {
	rusalphabetRune := []rune(rusalphabet)
	if node.leftson == nil && node.rightson != nil || node.leftson != nil && node.rightson == nil {
		fmt.Println("UNBALANCED!!!")
		os.Exit(-2)
	}
	if node.leftson == nil {
		index := findRuneInSlice(rusalphabetRune, node.symbol)
		codingTable.avgWordLen += float64(len(code)) * realProbs[index]
		codingTable.entropy -= node.probability * math.Log2(node.probability)
		codingTable.table[node.symbol] = code
		totalProb += node.probability
	} else {
		dfs(node.leftson, codingTable, code+"0", realProbs)
		dfs(node.rightson, codingTable, code+"1", realProbs)
	}
}

func dfsPair(node *NodeString, codingTable *CodingTableString, code string) {
	if node.leftson == nil && node.rightson != nil || node.leftson != nil && node.rightson == nil {
		fmt.Println("UNBALANCED!!!")
		os.Exit(-2)
	}
	if node.leftson == nil {
		codingTable.avgWordLen += float64(len(code)) * node.probability
		codingTable.entropy -= node.probability * math.Log2(node.probability)
		codingTable.table[node.symbol] = code
		totalProb += node.probability
	} else {
		dfsPair(node.leftson, codingTable, code+"0")
		dfsPair(node.rightson, codingTable, code+"1")
	}
}

func findRuneInSlice(slice []rune, symbol rune) int {
	for i, sliceSymbol := range slice {
		if sliceSymbol == symbol {
			return i
		}
	}
	return -1
}

func buildPairs() map[string]float64 {
	pairs := make(map[string]float64)
	for _, s1 := range rusalphabet {
		for _, s2 := range rusalphabet {
			pairs[string([]rune{s1, s2})] = 0.0
		}
	}
	return pairs
}

func countPairProbabilities(pairs map[string]float64, text string) {
	runeText := []rune(text)
	textLength := float64(len(runeText)) / 2.0
	for len(runeText) > 1 {
		currentPair := string(runeText[:2])
		runeText = runeText[2:]
		if v, ok := pairs[currentPair]; !ok {
			pairs[currentPair] = 1.0
		} else {
			pairs[currentPair] = v + 1.0
		}
	}
	for k, v := range pairs {
		pairs[k] = v / textLength
	}
}

func convertPairStatsToNodes(pairs map[string]float64) []*NodeString {
	nodes := make([]*NodeString, 0, 2000)
	for k, v := range pairs {
		if v != 0.0 {
			nodes = append(nodes, &NodeString{
				symbol:      k,
				probability: v,
			})
		}
	}
	return nodes
}
