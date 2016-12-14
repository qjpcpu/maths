package fixnum

import (
	"bytes"
	"errors"
	"fmt"
	"maths/tablewriter"
)

var Debug bool = false

//  +-----------+----------+----------+-----------+-----------+
//	|           | PRO-1(5) | PRO-2(0) | PRO-3(25) | PRO-4(60) |
//	+-----------+----------+----------+-----------+-----------+
//	| SKU-1(20) |        1 |        0 |         5 |        14 |
//	| SKU-2(30) |        1 |        0 |         8 |        21 |
//	| SKU-3(40) |        3 |        0 |        12 |        25 |
//	+-----------+----------+----------+-----------+-----------+
type Table struct {
	Promotions   []int64 // 每列总的优惠额
	Skus         []int64 // 每行的sku总价
	maxPromotion map[Cell]int64
	Data         [][]int64 // 结果:行为sku,列为优惠项
	History      []string  // Debug模式有用
}

// 单元格索引
type Cell struct {
	SkuI int
	ProI int
}

func (tbl Table) Render() string {
	b := bytes.NewBuffer([]byte{})
	table := tablewriter.NewWriter(b)
	headers := []string{""}
	for i, pro := range tbl.Promotions {
		headers = append(headers, fmt.Sprintf("PRO-%v(%v)", i, pro))
	}
	table.SetHeader(headers)

	for i, sku := range tbl.Skus {
		row := []string{fmt.Sprintf("SKU-%v(%v)", i, sku)}
		for _, cell := range tbl.Data[i] {
			row = append(row, fmt.Sprintf("%v", cell))
		}
		table.Append(row)
	}
	table.Render()
	return b.String()
}

func checkParams(promotions []int64, skus []int64, max map[Cell]int64) error {
	if len(promotions) == 0 || len(skus) == 0 {
		return errors.New("优惠项/SKU总价列表为空")
	}
	var match int64 = 0
	for _, e := range promotions {
		match += e
		if e < 0 {
			return errors.New("优惠额度不能为负")
		}
	}
	for _, e := range skus {
		match -= e
		if e <= 0 {
			return errors.New("SKU总价必须为正")
		}
	}
	if match != 0 {
		return fmt.Errorf("优惠总额应等于SKU总价之和,目前参数偏差: %v", match)
	}
	var keyUseless []Cell
	for cell, val := range max {
		if cell.SkuI < 0 || cell.ProI >= len(promotions) {
			keyUseless = append(keyUseless, cell)
		}
		if val < 0 {
			return fmt.Errorf("%+v的阈值必须为正数", cell)
		}
	}
	if len(keyUseless) > 0 {
		for _, key := range keyUseless {
			delete(max, key)
		}
	}
	// check sku max
	// 如果某sku每个优惠项都设置的最大值,那么最大值的和必须大于等于sku价格
	for i, total := range skus {
		next := false
		var sum int64 = 0
		for j, _ := range promotions {
			if m, ok := max[Cell{SkuI: i, ProI: j}]; !ok {
				next = true
				break
			} else {
				sum += m
			}
		}
		if next {
			continue
		}
		if sum < total {
			return fmt.Errorf("sku[%v] all promotion[%v] can't match total price[%v]", i, sum, total)
		}
	}
	// check promotion max
	// 如果某一个优惠项对下每个sku都设置的最大值,那么最大值的和必须大于等于优惠项
	for j, pro := range promotions {
		next := false
		var sum int64 = 0
		for i := range skus {
			if m, ok := max[Cell{SkuI: i, ProI: j}]; !ok {
				next = true
				break
			} else {
				sum += m
			}
		}
		if next {
			continue
		}
		if sum < pro {
			return fmt.Errorf("promotion[%v] all sku[%v] can't match total promotion[%v]", j, sum, pro)
		}
	}
	return nil
}

// author: qujianping
// date: 2016-12-06
// promotions每一列总的优惠
// skus每个sku总价
// maxPromotion某个单元格最大能承受的优惠,置为0表示不承担优化
func Explode(promotions []int64, skus []int64, maxPromotion map[Cell]int64) (Table, error) {
	result := make([][]int64, len(skus))
	table := Table{
		Promotions:   promotions,
		Skus:         skus,
		Data:         result,
		maxPromotion: maxPromotion,
	}
	// check parameters
	if err := checkParams(promotions, skus, maxPromotion); err != nil {
		return table, err
	}

	if maxPromotion == nil {
		maxPromotion = make(map[Cell]int64)
	}
	promotionTable := make([][]int64, len(skus))
	for i := range skus {
		promotionTable[i] = make([]int64, len(promotions))
		for j := range promotions {
			if val, ok := maxPromotion[Cell{SkuI: i, ProI: j}]; ok && val == 0 {
				promotionTable[i][j] = val
			} else {
				promotionTable[i][j] = promotions[j]
			}
		}
	}
	// 在爆破等价SKU时可能会有精度累积到最后一个优惠项的情形发生,这种情形可以进行一下优化
	canOpt := len(maxPromotion) == 0
	var cutPromotion []int64
	if canOpt {
		sku1 := skus[0]
		for _, v := range skus {
			if v != sku1 {
				canOpt = false
				break
			}
		}
		if canOpt {
			cutPromotion = make([]int64, len(promotions))
			copy(cutPromotion, promotions)
		}
	}
	var err error
	for i := 0; i < len(skus); i++ {
		if canOpt {
			for j, v := range cutPromotion {
				if v <= 0 {
					promotionTable[i][j] = 0
				}
			}
		}
		if result[i], err = DispatchByWeight(skus[i], promotionTable[i]); err != nil {
			return table, err
		}
		if canOpt {
			for j, v := range result[i] {
				cutPromotion[j] -= v
			}
		}
	}
	if ok, err := table.adjustColumn(); err != nil {
		return table, err
	} else if ok {
		if ok2, err2 := table.adjustMatrix(); err2 != nil {
			return table, err2
		} else if ok2 {
			return table, nil
		}
	}
	return table, errors.New("can't make it after 1000 times")
}

// 该步调平仅保证在有sku不参与某些优惠前提下, 横竖追平,但不保证某sku分得的某项优惠超出自身最大值限制
func (tbl *Table) adjustColumn() (bool, error) {
	if Debug {
		tbl.History = append(tbl.History, "初始分布:\n"+tbl.Render())
	}
	data := tbl.Data
	promotions := tbl.Promotions
	maxPromotion := tbl.maxPromotion

	diff := make([]int64, len(promotions))
	for j := 0; j < len(promotions); j++ {
		diff[j] = promotions[j]
		for i := 0; i < len(data); i++ {
			diff[j] -= data[i][j]
		}
	}
	omitProColumnIndex := make(map[int]bool)
	for len(omitProColumnIndex) <= len(promotions)-1 {
		// 互补列索引
		iless, imore := -1, -1
		for i, d := range diff {
			omit := omitProColumnIndex[i]
			if d > 0 && !omit {
				if iless == -1 || diff[iless] < d {
					iless = i
				}
			} else if d < 0 {
				if imore == -1 || diff[imore] < d {
					imore = i
				}
			}
		}
		if iless == -1 && imore == -1 {
			return true, nil
		}
		if iless == -1 || imore == -1 {
			// 无法找到互补列
			return false, errors.New("can't make it: complementary columns not found")
		}
		var val int64 = 0
		if diff[iless]+diff[imore] > 0 {
			val = -diff[imore]
		} else {
			val = diff[iless]
		}
		var max int64 = 0
		// 要找平的sku索引
		isku := -1
		for i := 0; i < len(data); i++ {
			disl, dism := false, false
			if d, ok := maxPromotion[Cell{SkuI: i, ProI: iless}]; ok && d == 0 {
				disl = true
			}
			if d, ok := maxPromotion[Cell{SkuI: i, ProI: imore}]; ok && d == 0 {
				dism = true
			}
			if data[i][imore] > max && !disl && !dism {
				max = data[i][imore]
				isku = i
			}
		}
		if isku < 0 {
			omitProColumnIndex[iless] = true
			continue
		}
		if max < val {
			val = max
		}
		data[isku][imore] -= val
		data[isku][iless] += val
		diff[imore] += val
		diff[iless] -= val
		// 调整日志
		if Debug {
			tbl.History = append(tbl.History, fmt.Sprintf("SKU-%v PRO-(%v,%v)-调整值%v\n%s", isku, iless, imore, val, tbl.Render()))
		}
	}
	if len(omitProColumnIndex) >= len(promotions)-1 {
		return false, errors.New("can't find two promotions to adjust")
	}
	for _, d := range diff {
		if d != 0 {
			return false, nil
		}
	}
	return true, nil
}

// 本调平保证不超出具体项最大值,采用矩形置换法
// A(+1)  <--------------- B(-1)
//   |                       |
//   |                       |
// X(-1)  ---------------> C(+1)
func (tbl *Table) adjustMatrix() (bool, error) {
	data := tbl.Data
	maxVals := tbl.maxPromotion
	for {
		// 找到分配点超出自身最大值
		cell := Cell{SkuI: -1, ProI: -1}
		//超出值
		var beyoundVal int64
		for c, val := range maxVals {
			if val > 0 && data[c.SkuI][c.ProI] > val {
				cell = c
				beyoundVal = data[c.SkuI][c.ProI] - val
				break
			}
		}
		if cell.SkuI == -1 || cell.ProI == -1 {
			break
		}
		// 交换对角点
		exCell := Cell{SkuI: -1, ProI: -1}
		for i := 0; i < len(data); i++ {
			if i == cell.SkuI {
				continue
			}
			for j := 0; j < len(data[i]); j++ {
				if j == cell.ProI {
					continue
				}
				// 对角点B如果等于0无法进行置换
				// 1. 无法进行减法 2. 可能设置了最大值为0
				if data[i][j] == 0 {
					continue
				}
				// 检查交换矩形另外两个角A,C是否能完成置换
				// 1. 如果禁用了优惠v==0,if恒成立
				// 2. 如果设置了优惠最大值,当前分得优惠需小于该最大值,保证A,C点至少能+1
				// 3. 如果没设置最大值,A,C必然能完成置换(+val)
				if v, ok := maxVals[Cell{SkuI: i, ProI: cell.ProI}]; ok && data[i][cell.ProI] >= v {
					continue
				}
				if v, ok := maxVals[Cell{SkuI: cell.SkuI, ProI: j}]; ok && data[cell.SkuI][j] >= v {
					continue
				}
				exCell.SkuI = i
				exCell.ProI = j
				// 执行到这里表示交换矩形成立,修正最大交换值
				if data[i][j] < beyoundVal {
					beyoundVal = data[i][j]
				}
				if v, ok := maxVals[Cell{SkuI: i, ProI: cell.ProI}]; ok && (v-data[i][cell.ProI]) < beyoundVal {
					beyoundVal = v - data[i][cell.ProI]
				}
				if v, ok := maxVals[Cell{SkuI: cell.SkuI, ProI: j}]; ok && (v-data[cell.SkuI][j]) < beyoundVal {
					beyoundVal = v - data[cell.SkuI][j]
				}
				// 开始一次矩形置换
				data[cell.SkuI][cell.ProI] -= beyoundVal
				data[exCell.SkuI][exCell.ProI] -= beyoundVal
				data[cell.SkuI][exCell.ProI] += beyoundVal
				data[exCell.SkuI][cell.ProI] += beyoundVal
				// 调整日志
				if Debug {
					tbl.History = append(tbl.History, fmt.Sprintf("矩形调整%+v - %+v调整值%v\n%s", cell, exCell, beyoundVal, tbl.Render()))
				}
				break
			}
			if exCell.SkuI >= 0 && exCell.ProI >= 0 {
				break
			}
		}
		if exCell.SkuI == -1 || exCell.ProI == -1 {
			return false, fmt.Errorf("can't find matrix for exchange values match %+v", cell)
		}
	}
	return true, nil
}

func DispatchByWeight(total int64, weights []int64) ([]int64, error) {
	length := len(weights)
	if length == 0 {
		return nil, errors.New("无分配权重")
	}
	if total <= 0 {
		return nil, errors.New("分配值需大于0")
	}
	var denominator int64 = 0
	for _, w := range weights {
		if w < 0 {
			return nil, errors.New("分配权重不能为负数")
		}
		denominator += w
	}
	if denominator == 0 {
		return nil, errors.New("分配权重不能全为0")
	}
	result := make([]int64, length)
	var cut int64 = 0
	for i := 0; i < length-1; i++ {
		val := int64(float64(weights[i]) / float64(denominator) * float64(total))
		cut += val
		result[i] = val
	}
	result[length-1] = total - cut
	// 修正最后权重为0的值
	if weights[length-1] == 0 && result[length-1] > 0 {
		for i := range result {
			if weights[i] > 0 {
				result[i] += result[length-1]
				result[length-1] = 0
				break
			}
		}
	}
	return result, nil
}

type SkuGroup struct {
	Data  []int64
	Count int64
}

// 将一种SKU爆破为单件SKU优惠分配
//  +--------------+---------------+-------------+------------+
//	| SKU-10(4800) |          4757 |          12 |         31 |
//	+--------------+---------------+-------------+------------+
// 对应已经分得的SKU-10,如果该sku单价1200买了4个,那么对该sku爆破的结果是
// [1190 3 7] x 1 = 1200
// [1189 3 8] x 3 = 3600
func ExplodeSku(promotions []int64, skuTotal int64, skuCount int64) ([]SkuGroup, error) {
	if skuTotal <= 0 {
		return nil, errors.New("SKU总价必须为正")
	}
	if skuCount <= 0 {
		return nil, errors.New("SKU数量为0")
	}
	if skuTotal%skuCount != 0 {
		return nil, errors.New("can't get single sku price")
	}
	if len(promotions) == 0 {
		return nil, errors.New("无促销项")
	}
	match := skuTotal
	for _, p := range promotions {
		match -= p
	}
	if match != 0 {
		return nil, errors.New("total promotions!=skuTotal")
	}
	// shortcut
	if skuCount == 1 {
		return []SkuGroup{
			SkuGroup{
				Data:  promotions,
				Count: 1,
			},
		}, nil
	}
	skus := make([]int64, skuCount)
	singlePrice := skuTotal / skuCount
	for i := range skus {
		skus[i] = singlePrice
	}
	tbl, err := Explode(promotions, skus, nil)
	if err != nil {
		return nil, err
	}
	// merge sku
	memo := make(map[string]SkuGroup)
	for i, row := range tbl.Data {
		key := fmt.Sprintf("%v", row)
		if sg, ok := memo[key]; ok {
			sg.Count += 1
			memo[key] = sg
		} else {
			memo[key] = SkuGroup{
				Data:  tbl.Data[i],
				Count: 1,
			}
		}
	}
	var list []SkuGroup
	for _, sg := range memo {
		list = append(list, sg)
	}
	return list, nil
}
