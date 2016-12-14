package fixnum

import (
	"testing"
)

func BenchmarkExplode(b *testing.B) {
	promotions := []int64{144388, 1468, 944}
	skus := []int64{4800, 4800, 25800, 19800, 15800, 18800, 4800, 13800, 28800, 4800, 4800}
	max := map[Cell]int64{
		Cell{SkuI: 3, ProI: 2}:  0,
		Cell{SkuI: 8, ProI: 1}:  0,
		Cell{SkuI: 10, ProI: 1}: 12,
		Cell{SkuI: 7, ProI: 0}:  11388,
	}
	for i := 0; i < b.N; i++ {
		tbl, err := Explode(promotions, skus, max)
		if err != nil {
			b.Fatal(err)
		}
		for _, row := range tbl.Data {
			var total int64
			for _, v := range row {
				total += v
			}
			ExplodeSku(row, total, 4)
		}
	}
}

func TestExplode(t *testing.T) {
	explodeNum := func(pro, skus []int64, dis map[Cell]int64) {
		table, err := Explode(pro, skus, dis)
		if err != nil {
			t.Log("explode failed\n", table.Render())
			t.Fatal(err)
		}
		t.Log("\n", table.Render())
	}

	var promotions, skus []int64
	promotions = []int64{5, 12, 13, 60}

	skus = []int64{20, 40, 30}
	dis := make(map[Cell]int64)
	dis[Cell{SkuI: 0, ProI: 1}] = 0
	dis[Cell{SkuI: 1, ProI: 1}] = 0
	explodeNum(promotions, skus, dis)

	promotions = []int64{20}
	skus = []int64{5, 7, 8}
	dis = make(map[Cell]int64)
	explodeNum(promotions, skus, dis)

	promotions = []int64{5, 7, 8}
	skus = []int64{20}
	explodeNum(promotions, skus, nil)

	promotions = []int64{5, 12, 13, 60}
	skus = []int64{30, 30, 30}
	dis = make(map[Cell]int64)
	dis[Cell{SkuI: 0, ProI: 0}] = 0
	dis[Cell{SkuI: 1, ProI: 1}] = 0
	explodeNum(promotions, skus, dis)

	promotions = []int64{0, 8, 4, 18}
	skus = []int64{10, 10, 10}
	explodeNum(promotions, skus, nil)

	promotions = []int64{1000, 1000, 6890}
	skus = []int64{6900, 1990}
	dis = make(map[Cell]int64)
	dis[Cell{SkuI: 1, ProI: 1}] = 0
	t.Log("可用卡币")
	explodeNum(promotions, skus, dis)
	t.Log("不可用卡币")
	dis[Cell{SkuI: 1, ProI: 0}] = 0
	explodeNum(promotions, skus, dis)

}

func TestExplodeMulti(t *testing.T) {
	Debug = true
	var promotions, skus []int64
	explodeNum := func(pro, skus []int64, dis map[Cell]int64) {
		table, err := Explode(pro, skus, dis)
		if err != nil {
			t.Log("explode failed\n", table.Render())
			t.Fatal(err)
		}
		for _, h := range table.History {
			t.Log(h)
		}
		t.Logf("times:%v Final\n%s\n", len(table.History), table.Render())
		// 切分sku
		for i, sk := range skus {
			sku := table.Data[i]
			t.Log("切分sku:", sku)
			list, err := ExplodeSku(sku, sk, 100)
			if err != nil {
				t.Fatal(err)
			}
			for _, single := range list {
				t.Logf("%v x %v\n", single.Data, single.Count)
			}
		}
	}
	// order 308659296037157561730809
	// promotions: 144388, 1468, 944
	// sku: 141657751836299537700323 coin: 48 coupon: 30 balance: 0 user: 4722 total: 4800
	// sku: 234521120636300916515917 coin: 48 coupon: 30 balance: 0 user: 4722 total: 4800
	// sku: 242427288234593159795777 coin: 258 coupon: 165 balance: 0 user: 25377 total: 25800
	// sku: 279202322835359006281839 coin: 198 coupon: 127 balance: 0 user: 19475 total: 19800
	// sku: 285569205035359557138572 coin: 158 coupon: 101 balance: 0 user: 15541 total: 15800
	// sku: 326376247635360586382175 coin: 188 coupon: 120 balance: 0 user: 18492 total: 18800
	// sku: 365255984436301439767383 coin: 48 coupon: 30 balance: 0 user: 4722 total: 4800
	// sku: 417516447535520513091412 coin: 138 coupon: 88 balance: 0 user: 13574 total: 13800
	// sku: 59590949235522694731517 coin: 288 coupon: 193 balance: 0 user: 28319 total: 28800
	// sku: 6058281535888776171863 coin: 48 coupon: 30 balance: 0 user: 4722 total: 4800
	// sku: 83244962136318864318903 coin: 48 coupon: 30 balance: 0 user: 4722 total: 4800
	promotions = []int64{144388, 1468, 944}
	skus = []int64{4800, 4800, 25800, 19800, 15800, 18800, 4800, 13800, 28800, 4800, 4800}
	max := map[Cell]int64{
		Cell{SkuI: 3, ProI: 2}:  0,
		Cell{SkuI: 8, ProI: 1}:  0,
		Cell{SkuI: 10, ProI: 1}: 12,
		//		Cell{SkuI: 7, ProI: 0}:  11388,
	}
	explodeNum(promotions, skus, max)
}
