package gopapageno

type CombineFunc func()

func SweepCombiner() {

}

func ParallelCombiner() {

}

func MixedCombiner(parallelPasses int) CombineFunc {
	return func() {

	}
}
