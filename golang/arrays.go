package main

import (
	"fmt"
	"math/big"
)

func removeItem(slice *[]int, s int) {
	*slice = append((*slice)[:s], (*slice)[s+1:]...)
}

func identifyNonPrimes(ptr *[]int) {
	fmt.Println("Items count=",len(*ptr))
	for i := len(*ptr)-1; i >=0 ; i-- {
		if !(big.NewInt(int64((*ptr)[i])).ProbablyPrime(0)) {
			fmt.Println("item at index",i,"which is",(*ptr)[i], "is not prime")
			removeItem(ptr, i)
			fmt.Println("updated list",*ptr, "has length",len(*ptr))
		}
	}
}

func remove(slice []int, i int) []int {
	fmt.Println(slice[i:]) // [7 8 9]
	fmt.Println(slice[i+1:]) // [8 9]
	copy(slice[i:], slice[i+1:]) //func copy(dst, src []Type) int
	fmt.Println(slice[:]) // [5 6 8 9 9]
	return slice[:len(slice)-1]
}

func main() {
	// primes := []int{2, 3, 4, 6, 6, 7, 8, 9, 10, 11, 13}
	// fmt.Println(primes)
	// identifyNonPrimes(&primes)
	// fmt.Println(primes)
	s := []int{5, 6, 7, 8, 9}
	fmt.Println(remove(s, 2)) // "[5 6 8 9]"
}
