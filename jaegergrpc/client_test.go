package jaegergrpc

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"testing"
)

func TestGrpcClientExample(t *testing.T) {
	myServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_opentracing.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_opentracing.UnaryServerInterceptor(),
		)),
	)
	myServer.ServeHTTP(nil, nil)
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

type ListNode struct {
	Val  int
	Next *ListNode
}
type Arr struct {
	arr []*TreeNode
}

func (a *Arr) push(node *TreeNode) {
	a.arr = append(a.arr, node)
}
func (a *Arr) pop() *TreeNode {
	n := a.arr[0]
	a.arr = a.arr[1:]
	return n
}
func (a *Arr) isEmpty() bool {
	return len(a.arr) == 0
}
func isCompleteTree(root *TreeNode) bool {
	arr1 := Arr{make([]*TreeNode, 0)}
	arr2 := Arr{make([]*TreeNode, 0)}
	arr1.push(root)
	mNode := 0
	for !arr1.isEmpty() || !arr2.isEmpty() {
		if !arr1.isEmpty() {
			for !arr1.isEmpty() {
				n := arr1.pop()
				if n.Left == nil && n.Right == nil {
					continue
				}
				if n.Left != nil {
					arr2.push(n.Left)
				} else {
					mNode++
				}
				if n.Right != nil {
					if n.Left == nil {
						return false
					}
					arr2.push(n.Right)
				} else {
					mNode++
				}
			}
			continue
		}

		if !arr2.isEmpty() {
			for !arr2.isEmpty() {
				n := arr2.pop()
				if n.Left == nil && n.Right == nil {
					continue
				}
				if n.Left != nil {
					arr1.push(n.Left)
				} else {
					mNode++
				}
				if n.Right != nil {
					if n.Left == nil {
						return false
					}
					arr1.push(n.Right)
				} else {
					mNode++
				}
			}
			continue
		}
	}
	if mNode > 1 {
		return false
	}
	return true
}

func TestGet(t *testing.T) {
	n := &TreeNode{1,
		&TreeNode{2,
			&TreeNode{5, nil, nil},
			nil},
		&TreeNode{3,
			&TreeNode{7, nil, nil},
			&TreeNode{8, nil, nil}},
	}
	fmt.Println(isCompleteTree(n))
}
