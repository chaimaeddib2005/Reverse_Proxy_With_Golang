package main

import( "fmt" 
		"log"
		"net/http"
		"os"
)

func main(){
	if len(os.Args) < 2{
		log.Fatal("Usage: go run mock_backend.go <port>")		
	}
	port := os.Args[1]
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		response := fmt.Sprintf("Response from backend on post %s\n",port)
		response += fmt.Sprintf("Request Path: %s\n",r.URL.Path)
		response += fmt.Sprintf("Received request: %s %s",r.Method, r.URL.Path)
		w.Write([]byte(response))

	})
	address := ":"+port
	log.Printf("Mock backend server starting on %s", address)
	log.Fatal(http.ListenAndServe(address,nil))
} 