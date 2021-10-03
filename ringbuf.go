package main

import (
	"bytes"
	"errors"
)

// TODO: size 设置为 2^n 取模相当于 & (size-1)即可
// TODO: 加锁
var (
	ErrIsFull             = errors.New("ringbuf is full")
	ErrTooManyDataToWrite = errors.New("to many data to write")
	ErrIsEmpty            = errors.New("ringbuf is empty")
)

// w 不存储实际数据
type ringbuf struct {
	w    int
	r    int
	size int
	buf  []byte
}

func New(size int) *ringbuf {
	return &ringbuf{
		size: size + 1,
		buf:  make([]byte, size+1),
	}
}

func (r *ringbuf) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	n, err = r.read(p)
	return
}

func (r *ringbuf) read(p []byte) (n int, err error) {
	if r.isEmpty() {
		return 0, ErrIsEmpty
	}
	// w 指针在 r 指针后面，说明 w 指针还没有到 buf 最后
	// 还可以继续写
	if r.w > r.r {
		n = r.w - r.r

		if n > len(p) {
			n = len(p)
		}
		copy(p, r.buf[r.r:r.r+n])
		// r 在write后面，+ n 也不会超过r.w，不需要取余操作
		r.r += n
		return
	}

	// w 指针在 r 指针前面
	n = r.r - r.w + r.size
	if n > len(p) {
		n = len(p)
	}
	// 顺时针读即可，不用绕
	if r.r+n <= r.size {
		copy(p, r.buf[r.r:r.r+n])
	} else {
		// buf 后面写了一段数据
		n1 := r.size - r.r
		copy(p, r.buf[r.r:])
		// 然后绕到 buf 前面写数据
		n2 := n - n1 //  n - (r.size - r.r)
		copy(p[n1:], r.buf[:n2])
	}
	// 有可能 r.r + n >= size
	r.r = (r.r + n) % r.size
	return
}

func (r *ringbuf) Write(p []byte) (n int, err error) {
	n, err = r.write(p)
	return
}

func (r *ringbuf) write(p []byte) (n int, err error) {
	if r.IsFull() {
		return 0, ErrIsFull
	}

	// 注意需要-1，因为 w 实际上不存储数据
	available := (r.size - r.w + r.r - 1) % r.size

	if available < len(p) {
		return 0, ErrTooManyDataToWrite
	}

	n = len(p)

	// w 指针在 r 指针前面
	if r.w >= r.r {
		// w 指针后面的空间是否足够
		n1 := r.size - r.w
		// 足够的话直接写 buf 后面就好
		if n1 >= n {
			copy(r.buf[r.w:], p)
			r.w += n

			// 否则还需要在 buf 前面写一段数据
		} else {
			copy(r.buf[r.w:], p[:n1])
			// buf 前面还需要写的数据数量
			n2 := n - n1
			copy(r.buf[:n2], p[n1:])
			r.w = n2
		}
	} else {
		copy(r.buf[r.w:], p)
		r.w += n
	}
	if r.w == r.size {
		r.w = 0
	}

	return
}

func (r *ringbuf) isEmpty() bool {
	return r.w == r.r
}

func (r *ringbuf) IsFull() bool {
	return (r.w+1)%r.size == r.r
}

func (r *ringbuf) Reset() {
	r.r = 0
	r.w = 0
}

func (r *ringbuf) Bytes() []byte {
	buf := bytes.Buffer{}
	for i := r.r; i != r.w; i = (i + 1) % r.size {
		buf.WriteByte(r.buf[i])
	}
	return buf.Bytes()
}

func (r *ringbuf) String() string {
	buf := bytes.Buffer{}
	for i := r.r; i != r.w; i = (i + 1) % r.size {
		buf.WriteByte(r.buf[i])
	}
	return buf.String()
}
