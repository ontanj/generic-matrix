package genmatrix

import (
    "fmt"
)

type Matrix struct {
    values []interface{}
    Rows, Cols int
    Space Space
}

// create a new Matrix with the given size and data acting in space
func NewMatrix(rows, cols int, data []interface{}, space Space) (m Matrix, err error) {
    if data == nil {
        data = make([]interface{}, rows*cols)
    } else if rows * cols != len(data) {
        err = fmt.Errorf("Data structure not matching matrix size: %d x %d != %d", rows, cols, len(data))
        return
    }
    if space == nil {
        err = fmt.Errorf("space can't be nil")
        return
    }
    m.values = data
    m.Rows = rows
    m.Cols = cols
    m.Space = space
    return
}

// get value at (row, col), where first row/col is 0.
func (m Matrix) At(row, col int) (interface{}, error) {
    if row >= m.Rows || col >= m.Cols || row < 0 || col < 0{
        return nil, fmt.Errorf("Index out of bounds: (%d, %d)", row, col)
    }
    valueIndex := m.Cols*row + col
    return m.values[valueIndex], nil
}

// set value at (row, col), where first row/col is 0.
func (m Matrix) Set(row, col int, value interface{}) error {
    if row >= m.Rows || col >= m.Cols || row < 0 || col < 0 {
        return fmt.Errorf("Index out of bounds: (%d, %d)", row, col)
    }
    m.values[m.Cols*row + col] = value
    return nil
}

// multiply a * b
// also handles multiplication of scalar * non-scalar matrices and vice versa
// if a and b are non-scalar in different spaces, the space of a is used
func (a Matrix) Multiply(b Matrix) (c Matrix, err error) {
    if a.Cols != b.Rows {
        err = fmt.Errorf("matrices a and b are not compatible")
        return
    }
    cRows, cCols := a.Rows, b.Cols
    values := make([]interface{}, cRows*cCols)
    var r, a_val, b_val interface{}
    var space Space
    for i := 0; i < cRows; i += 1 {
        for j := 0; j < cCols; j += 1 {
            var sum interface{}
            for k := 0; k < a.Cols; k += 1 {
                a_val, err = a.At(i, k)
                b_val, err = b.At(k, j)
                if err != nil {return}
                if a.Space.Scalarspace() {
                    space = b.Space
                    r, err = b.Space.Scale(b_val, a_val)
                    if err != nil {return a, err}
                    if sum == nil {
                        sum = r
                    } else {
                        sum, err = b.Space.Add(r, sum)
                        if err != nil {return a, err}
                    }
                } else {
                    space = a.Space
                    if b.Space.Scalarspace() {
                        r, err = a.Space.Scale(a_val, b_val)
                        if err != nil {return a, err}
                    } else {
                        r, err = a.Space.Multiply(a_val, b_val)
                        if err != nil {return a, err}
                    }
                    if sum == nil {
                        sum = r
                    } else {
                        sum, err = a.Space.Add(r, sum)
                        if err != nil {return a, err}
                    }
                }
            }
            values[i*cCols+j] = sum
            sum = nil
        }
    }
    return NewMatrix(cRows, cCols, values, space)
}

// multiplication of a by a scalar
// assumes matrix and factor is in same space, otherwise use Scale
func (a Matrix) MultiplyScalar(scalar interface{}) (Matrix, error) {
    return scalarMultiplication(a.Space.Multiply, a, scalar)
}

// scale a according to scalar
// to be used if factor is in a scalar space wile a is not
func (a Matrix) Scale(factor interface{}) (Matrix, error) {
    return scalarMultiplication(a.Space.Scale, a, factor)
}

func scalarMultiplication(mulfunc func(interface{}, interface{}) (interface{}, error), a Matrix, b interface{}) (Matrix, error) {
    c_vals := make([]interface{}, len(a.values))
    var err error
    for i := range c_vals {
        c_vals[i], err = mulfunc(a.values[i], b)
        if err != nil {return a, err}
    }
    return NewMatrix(a.Rows, a.Cols, c_vals, a.Space)
}

// matrix addition
func (a Matrix) Add(b Matrix) (c Matrix, err error) {
    if a.Rows != b.Rows || a.Cols != b.Cols {
        err = fmt.Errorf("dimension mismatch in addition: %d x %d != %d x %d", a.Rows, a.Cols, b.Rows, b.Cols)
        return
    }
    c_vals := make([]interface{}, len(a.values))
    for i := range c_vals {
        c_vals[i], err = a.Space.Add(a.values[i], b.values[i])
        if err != nil {return a, err}
    }
    return NewMatrix(a.Rows, a.Cols, c_vals, a.Space)
}

// matrix subtraction
func (a Matrix) Subtract(b Matrix) (c Matrix, err error) {
    if a.Rows != b.Rows || a.Cols != b.Cols {
        err = fmt.Errorf("dimension mismatch in subtraction: %d x %d != %d x %d", a.Rows, a.Cols, b.Rows, b.Cols)
        return
    }
    c_vals := make([]interface{}, len(a.values))
    for i := range c_vals {
        c_vals[i], err = a.Space.Subtract(a.values[i], b.values[i])
        if err != nil {return a, err}
    }
    return NewMatrix(a.Rows, a.Cols, c_vals, a.Space)
}

// concatenate matrices as A|B
func (a Matrix) Concatenate(b Matrix) (Matrix, error) {
    if a.Rows != b.Rows {
        return Matrix{}, fmt.Errorf("matrices not compatible for concatenation, a has %d rows while b has %d rows", a.Rows, b.Rows)
    }
    vals := make([]interface{}, 0, (a.Cols + b.Cols) * a.Rows)
    for i := 0; i < a.Rows; i += 1 {
        vals = append(vals, a.values[i*a.Cols:(i+1)*a.Cols]...)
        vals = append(vals, b.values[i*b.Cols:(i+1)*b.Cols]...)
    }
    return NewMatrix(a.Rows, a.Cols + b.Cols, vals, a.Space)
}

// create a new matrix from last k columns of a
func (a Matrix) CropHorizontally(k int) Matrix {
    vals := make([]interface{}, 0, k*a.Rows)
    d := a.Cols - k
    for i := 0; i < a.Rows; i += 1 {
        vals = append(vals, a.values[i*a.Cols+d:(i+1)*a.Cols]...)
    }
    c, _ := NewMatrix(a.Rows, k, vals, a.Space)
    return c
}

// apply function f to all matrix elements
func (a Matrix) Apply(f func(interface{}) (interface{}, error)) (b Matrix, err error) {
    b_vals := make([]interface{}, len(a.values))
    for i, v := range a.values {
        b_vals[i], err = f(v)
        if err != nil {return}
    }
    return NewMatrix(a.Rows, a.Cols, b_vals, a.Space)
}