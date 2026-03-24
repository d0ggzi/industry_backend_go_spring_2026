package main

type Stats struct {
	Count int
	Sum   int64
	Min   int64
	Max   int64
}

func Calc(nums []int64) Stats {
	if len(nums) == 0 {
		return Stats{}
	}

	res := Stats{
		Count: 0,
		Sum:   0,
		Min:   nums[0],
		Max:   nums[0],
	}

	// 3. Проходим по всем элементам один раз
	for _, v := range nums {
		res.Sum += v
		res.Count += 1

		if v < res.Min {
			res.Min = v
		}

		if v > res.Max {
			res.Max = v
		}
	}

	return res
}
