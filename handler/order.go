package handler

import (
	"fmt"
	"net/http"

	"github.com/LinerGit/godev/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Order created")
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Order details")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Order updated")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Order deleted")
}

func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Order details")
}
