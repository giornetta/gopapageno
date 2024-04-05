package xpath

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRpeInnerTest(t *testing.T) {
	t.Run(`rpeInnerTest.String()`, func(t *testing.T) {
		t.Run(`rpeInnerTest.behindAncestorAxis==true`, func(t *testing.T) {
			elementTest := newElementTest("a", nil, nil)
			rpeInnerTest := &rpeInnerTestImpl{
				udpeTest:           elementTest,
				behindAncestorAxis: true,
			}

			want := fmt.Sprintf(`\\%v`, elementTest)
			if got := rpeInnerTest.String(); got != want {
				t.Errorf(`fpeInnerTest.String()=%v | want %v`, got, want)
			}
		})

		t.Run(`rpeInnerTest.behindAncestorAxis==false`, func(t *testing.T) {
			elementTest := newElementTest("a", nil, nil)
			rpeInnerTest := &rpeInnerTestImpl{
				udpeTest: elementTest,
			}

			t.Run(`rpeInnerTest.isEntry==false`, func(t *testing.T) {
				want := fmt.Sprintf(`\%v`, elementTest)

				if got := rpeInnerTest.String(); got != want {
					t.Errorf(`rpeInnerTest.String()=%v | want %v`, got, want)
				}
			})

			t.Run(`rpeInnerTest.isEntry==true`, func(t *testing.T) {
				rpeInnerTest.isEntry = true

				want := fmt.Sprintf(`%v`, elementTest)
				if got := rpeInnerTest.String(); got != want {
					t.Errorf(`rpeInnerTest.String()=%v | want %v`, got, want)
				}
			})
		})
	})
}

func TestRpeBuilder(t *testing.T) {
	t.Run(`rpeBuilder.addUdpeTest(t)`, func(t *testing.T) {
		t.Run(`rpeBuilder.addUdpeTest(t)=false if rpeBuilder expecting an axis`, func(t *testing.T) {
			rpeBuilder := new(rpeBuilderImpl)

			if rpeBuilder.addUdpeTest(newElementTest("a", nil, nil)) {
				t.Error(`rpeBuilder.addUdpeTest(t)=true | want false`)
			}
		})

		t.Run(`rpeBuilder(addUdpeTest(t)=true if rpeBuilder expecting an udpeTest`, func(t *testing.T) {
			rpeBuilder := new(rpeBuilderImpl)
			rpeBuilder.state = expectRpeUdpeTest

			udpeTest := newElementTest("a", nil, nil)
			if !rpeBuilder.addUdpeTest(udpeTest) {
				t.Error(`rpeBuilder.addUdpeTest(t)=false | want true`)
			}
			if rpeBuilder.state != expectRpeAxis {
				t.Error(`rpeBuilder.addUdpeTest(t) does NOT update rpeBuilder.state correctly`)
			}
			expectedCurrentInnerRpeInnerTest := &rpeInnerTestImpl{
				udpeTest:         udpeTest,
				nextRpeInnerTest: nil,
			}

			if got := rpeBuilder.currentInnerTest; !reflect.DeepEqual(*got, *expectedCurrentInnerRpeInnerTest) {
				t.Error(`rpeBuilder.addUdpeTest(t) does NOT update rpeBuilder.currentInnerTest correctly`)
			}
		})
	})

	t.Run(`rpeBuilder.addAxis(a)`, func(t *testing.T) {
		t.Run(`rpeBuilder.addAxis(a)=false if rpeBuilder expecting an udpeTest`, func(t *testing.T) {
			rpeBuilder := new(rpeBuilderImpl)
			rpeBuilder.state = expectRpeUdpeTest

			if rpeBuilder.addAxis(child) {
				t.Error(`rpeBuilder.addAxis(a)=true | want false`)
			}
		})

		t.Run(`rpeBuilder.addAxis(a)=true if rpeBuilder expecting an udpeTest`, func(t *testing.T) {
			t.Run(`rpeBuilder.currentInnerTest=nil`, func(t *testing.T) {
				t.Run(`a=parent`, func(t *testing.T) {
					rpeBuilder := new(rpeBuilderImpl)

					if !rpeBuilder.addAxis(parent) {
						t.Error(`rpeBuilder.addAxis(parent)=false | want trye`)
					}

					if rpeBuilder.state != expectRpeUdpeTest {
						t.Error(`rpeBuilder.addAxis(parent) does NOT update rpeBuilder.state correclty`)
					}
				})

				t.Run(`a=ancestorOrSelf`, func(t *testing.T) {
					rpeBuilder := new(rpeBuilderImpl)

					if !rpeBuilder.addAxis(ancestorOrSelf) {
						t.Error(`rpeBuilder.addAxis(ancestorOrSelf)= false | want true`)
					}

					if rpeBuilder.state != expectRpeUdpeTest {
						t.Error(`rpeBuilder.addAxis(ancestorOrSelf) does NOT update rpeBuilder.state correctly`)
					}
				})
			})

			t.Run(`rpeBuilder.currentInnerTest !=nil`, func(t *testing.T) {
				t.Run(`a=parent`, func(t *testing.T) {
					rpeBuilder := new(rpeBuilderImpl)
					rpeInnerTest := new(rpeInnerTestImpl)
					rpeBuilder.currentInnerTest = rpeInnerTest

					if !rpeBuilder.addAxis(parent) {
						t.Error(`rpeBuilder.addAxis(parent)=false | want true`)
					}

					if rpeBuilder.state != expectRpeUdpeTest {
						t.Error(`rpeBuilder.addAxis(parent) does NOT update rpeBuilder.state correctly`)
					}

					if !reflect.DeepEqual(*rpeBuilder.currentInnerTest, *rpeInnerTest) {
						t.Error(`rpeBuilder.addAxis(parent) should NOT update rpeBuilder.currentInnerTest`)
					}
				})

				t.Run(`a=ancestorOrSelf`, func(t *testing.T) {
					rpeBuilder := new(rpeBuilderImpl)
					rpeBuilder.currentInnerTest = new(rpeInnerTestImpl)

					if !rpeBuilder.addAxis(ancestorOrSelf) {
						t.Error(`rpebuilder.addAxis(ancestorOrSelf)=false | want true`)
					}

					if rpeBuilder.state != expectRpeUdpeTest {
						t.Error(`rpeBuilder.addAxis(ancestorOrSelf) does NOT update rpeBuilder.state correctly`)
					}

					if currentInnerTest := rpeBuilder.currentInnerTest; !currentInnerTest.behindAncestorAxis {
						t.Error(`rpeBuilder.addAxis(ancestorOrSelf) does NOT update rpeBuilder.currentInnerState correctly`)
					}
				})
			})
		})
	})

	t.Run(`rpeBuilder.end()`, func(t *testing.T) {
		t.Run(`rpeBuilder.end()=nil if NO udpeTest has been added`, func(t *testing.T) {
			rpeBuilder := new(rpeBuilderImpl)

			if rpeBuilder.end() != nil {
				t.Errorf(`rpeBuilder.end() !nil | want nil`)
			}
		})

		t.Run(`rpeBuilder.end() !=nil if at least one udpeTest has been added`, func(t *testing.T) {
			rpeBuilder := new(rpeBuilderImpl)
			rpeBuilder.addAxis(parent)
			rpeBuilder.addUdpeTest(newElementTest("a", nil, nil))

			rpe := rpeBuilder.end().(*rpeImpl)
			if rpe == nil {
				t.Error(`rpeBuilder.end()=nil | want rpe`)
			}

			if !rpe.entryTest.isEntry {
				t.Error(`rpeBuilder.end().entryTest.isEntry=false | want true`)
			}
		})
	})
}

func TestRpe(t *testing.T) {
	t.Run(`rpe.String()`, func(t *testing.T) {
		t.Run(`rpe=a\b\`, func(t *testing.T) {
			rpeBuilder := newRpeBuilder()
			rpeBuilder.addAxis(parent)
			rpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			rpeBuilder.addAxis(parent)
			rpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			rpe := rpeBuilder.end()
			want := `a\b`

			if got := rpe.String(); got != want {
				t.Errorf(`rpe.String()=%v | want %v`, got, want)
			}
		})

		t.Run(`rpe=a\\b`, func(t *testing.T) {
			rpeBuilder := newRpeBuilder()
			rpeBuilder.addAxis(parent)
			rpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			rpeBuilder.addAxis(ancestorOrSelf)
			rpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			rpe := rpeBuilder.end()
			want := `a\\b`

			if got := rpe.String(); got != want {
				t.Errorf(`rpe.String()=%v | want %v`, got, want)
			}
		})
	})
}

// Test integraton that makes Algorithm 2
func TestRpeIntegration(t *testing.T) {
	t.Run(`Phase 1`, func(t *testing.T) {
		t.Run(`1. γ = α<b>X</b>`, func(t *testing.T) {
			reducedElement := newElement("b", []*Attribute{NewAttribute("key", "value")}, nil)

			t.Run(`2. 6.`, func(t *testing.T) {
				//2.
				rpeBuilder1 := newRpeBuilder()
				rpeBuilder1.addAxis(parent)
				rpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder1.addAxis(parent)
				rpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))

				rpeBuilder2 := newRpeBuilder()
				rpeBuilder2.addAxis(parent)
				rpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder2.addAxis(parent)
				rpeBuilder2.addUdpeTest(newElementTest("*", nil, nil))

				rpeBuilder3 := newRpeBuilder()
				rpeBuilder3.addAxis(parent)
				rpeBuilder3.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder3.addAxis(ancestorOrSelf)
				rpeBuilder3.addUdpeTest(newElementTest("*", nil, nil))

				//6.
				rpeBuilder4 := newRpeBuilder()
				rpeBuilder4.addAxis(parent)
				rpeBuilder4.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder4.addAxis(ancestorOrSelf)
				rpeBuilder4.addUdpeTest(newElementTest("b", nil, nil))

				var tests = []struct {
					rpeBuilder                           rpeBuilder
					expectedRpePathPatternReprAfterMatch string
				}{
					{
						rpeBuilder:                           rpeBuilder1,
						expectedRpePathPatternReprAfterMatch: `a\ε`,
					},
					{
						rpeBuilder:                           rpeBuilder2,
						expectedRpePathPatternReprAfterMatch: `a\ε`,
					},
					{
						rpeBuilder:                           rpeBuilder3,
						expectedRpePathPatternReprAfterMatch: `\\a\ε`,
					},
					{
						rpeBuilder:                           rpeBuilder4,
						expectedRpePathPatternReprAfterMatch: `\\a\ε`,
					},
				}

				for _, test := range tests {
					rpe := test.rpeBuilder.end()
					t.Run(`rpe=`+rpe.String(), func(t *testing.T) {
						entryPoint := rpe.entryPoint()

						_, newPathPattern, ok := entryPoint.matchWithReductionOf(reducedElement, true)

						if !ok {
							t.Error(`rpe entry point does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`entry point returns a new path pattern when it should NOT`)
						}

						if entryPoint.String() != test.expectedRpePathPatternReprAfterMatch {
							t.Error(`rpe entry point does NOT update correctly`)
						}
					})
				}
			})

			t.Run(`10.`, func(t *testing.T) {
				rpeBuilder1 := newRpeBuilder()
				rpeBuilder1.addAxis(parent)
				rpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder1.addAxis(parent)
				rpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))
				rpeBuilder1.addAxis(ancestorOrSelf)

				rpeBuilder2 := newRpeBuilder()
				rpeBuilder2.addAxis(parent)
				rpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
				rpeBuilder2.addAxis(parent)
				rpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))
				rpeBuilder2.addAxis(ancestorOrSelf)

				rpeBuilder3 := newRpeBuilder()
				rpeBuilder3.addAxis(parent)
				rpeBuilder3.addUdpeTest(newElementTest("b", nil, nil))
				rpeBuilder3.addAxis(ancestorOrSelf)

				var tests = []struct {
					rpeBuilder            rpeBuilder
					returnsNewPathPattern bool
					newPathPatternRepr    string
				}{
					{
						rpeBuilder:            rpeBuilder1,
						returnsNewPathPattern: true,
						newPathPatternRepr:    `a\ε`,
					},
					{
						rpeBuilder:            rpeBuilder2,
						returnsNewPathPattern: false,
					},
					{
						rpeBuilder:            rpeBuilder3,
						returnsNewPathPattern: true,
						newPathPatternRepr:    `ε`,
					},
				}

				for _, test := range tests {
					rpe := test.rpeBuilder.end()
					t.Run(`rpe=`+rpe.String(), func(t *testing.T) {
						entryPoint := rpe.entryPoint()
						entryPointReprBeforeMatch := entryPoint.String()

						_, newPathPattern, ok := entryPoint.matchWithReductionOf(reducedElement, true)

						if !ok {
							t.Error(`rpe entry point does NOT match a compatible Reduction`)
						}

						if entryPoint.String() != entryPointReprBeforeMatch {
							t.Error(`rpe entry point does NOT update correctly`)
						}

						if test.returnsNewPathPattern {
							if newPathPattern == nil || newPathPattern.String() != test.newPathPatternRepr {
								t.Error(`rpe entry point returns wrong path pattern`)
							}
						} else {
							if newPathPattern != nil {
								t.Error(`rpe entry point returns a new path pattern when it should NOT`)
							}
						}
					})
				}
			})
		})
	})

	t.Run(`Phase 2`, func(t *testing.T) {
		entryPointConsumingReduction := newElement("c", nil, nil)
		positiveReduction := newElement("b", nil, nil)
		negativeReduction := newElement("z", nil, nil)

		t.Run(`13.`, func(t *testing.T) {
			rpeBuilder1 := newRpeBuilder()
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("b", nil, nil)) // <- actual tested match
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			rpeBuilder2 := newRpeBuilder()
			rpeBuilder2.addAxis(parent)
			rpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder2.addAxis(ancestorOrSelf)
			rpeBuilder2.addUdpeTest(newElementTest("b", nil, nil)) // <- actual tested match
			rpeBuilder2.addAxis(parent)
			rpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				rpeBuilder                           rpeBuilder
				expectedRpePathPatternReprAfterMatch string
			}{
				{
					rpeBuilder:                           rpeBuilder1,
					expectedRpePathPatternReprAfterMatch: `a\ε`,
				},
				{
					rpeBuilder:                           rpeBuilder2,
					expectedRpePathPatternReprAfterMatch: `\\a\ε`,
				},
			}

			for _, test := range tests {
				rpe := test.rpeBuilder.end()
				rpePathPattern := rpe.entryPoint()
				rpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)

				t.Run(`rpe path pattern=`+rpePathPattern.String(), func(t *testing.T) {
					t.Run(`negative matching`, func(t *testing.T) {
						_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(negativeReduction, true)
						if ok {
							t.Error(`rpe path pattern matches an incompatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
						}
					})

					t.Run(`positive matching`, func(t *testing.T) {
						_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(positiveReduction, true)

						if !ok {
							t.Error(`rpe path pattern does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
						}

						if rpePathPattern.String() != test.expectedRpePathPatternReprAfterMatch {
							t.Error(`rpe path pattern does NOT update correctly`)
						}
					})
				})
			}
		})

		t.Run(`16.`, func(t *testing.T) {
			rpeBuilder1 := newRpeBuilder()
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))
			rpeBuilder1.addAxis(ancestorOrSelf) //<- actual tested match
			rpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			rpeBuilder2 := newRpeBuilder()
			rpeBuilder2.addAxis(parent)
			rpeBuilder2.addUdpeTest(newElementTest("b", nil, nil))
			rpeBuilder2.addAxis(ancestorOrSelf)
			rpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				rpeBuilder                           rpeBuilder
				expectedRpePathPatternReprAfterMatch string
				newPathPatternRepr                   string
			}{
				{
					rpeBuilder:                           rpeBuilder1,
					expectedRpePathPatternReprAfterMatch: `\\b\a\ε`,
					newPathPatternRepr:                   `a\ε`,
				},
				{
					rpeBuilder:                           rpeBuilder2,
					expectedRpePathPatternReprAfterMatch: `\\b\ε`,
					newPathPatternRepr:                   `ε`,
				},
			}

			for _, test := range tests {
				rpe := test.rpeBuilder.end()
				rpePathPattern := rpe.entryPoint()
				rpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)

				t.Run(`rpe path pattern=`+rpePathPattern.String(), func(t *testing.T) {
					t.Run(`18.`, func(t *testing.T) {
						_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(negativeReduction, true)

						if !ok {
							t.Error(`rpe path pattern does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
						}

						if rpePathPattern.String() != test.expectedRpePathPatternReprAfterMatch {
							t.Error(`rpe path pattern does NOT update correctly after match`)
						}
					})

					t.Run(`17.`, func(t *testing.T) {
						_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(positiveReduction, true)

						if !ok {
							t.Error(`rpe path pattern does NOT match a compatible Reduction`)
						}

						if rpePathPattern.String() != test.expectedRpePathPatternReprAfterMatch {
							t.Error(`rpe path pattern does NOT update correctly after match`)
						}

						if newPathPattern == nil || newPathPattern.String() != test.newPathPatternRepr {
							t.Error(`rpe path pattern returns wrong new path pattern`)
						}
					})
				})
			}
		})

		t.Run(`19.`, func(t *testing.T) {
			const expectedRpePathPatternReprAfterMatch = `\\a\ε`

			rpeBuilder := newRpeBuilder()
			rpeBuilder.addAxis(parent)
			rpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder.addAxis(ancestorOrSelf)
			rpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			rpeBuilder.addAxis(ancestorOrSelf) //<- actual tested match
			rpeBuilder.addUdpeTest(newElementTest("c", nil, nil))

			rpe := rpeBuilder.end()
			rpePathPattern := rpe.entryPoint()
			rpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)
			rpePathPatternReprBeforeMatch := rpePathPattern.String()

			t.Run(`rpe path pattern=`+rpePathPattern.String(), func(t *testing.T) {
				t.Run(`21.`, func(t *testing.T) {
					_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(negativeReduction, true)

					if !ok {
						t.Error(`rpe path pattern doesn NOT match a compatible Reduction`)
					}

					if newPathPattern != nil {
						t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
					}

					if rpePathPattern.String() != rpePathPatternReprBeforeMatch {
						t.Error(`rpe path pattern does NOT update correctly`)
					}
				})

				t.Run(`21.`, func(t *testing.T) {
					_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(positiveReduction, true)

					if !ok {
						t.Error(`rpe path pattern doesn NOT match a compatible Reduction`)
					}

					if newPathPattern != nil {
						t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
					}

					if rpePathPattern.String() != expectedRpePathPatternReprAfterMatch {
						t.Error(`rpe path pattern does NOT update correctly`)
					}
				})
			})
		})

		t.Run(`22.`, func(t *testing.T) {
			rpeBuilder1 := newRpeBuilder()
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("*", nil, nil)) // <- actual tested match
			rpeBuilder1.addAxis(parent)
			rpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			rpeBuilder2 := newRpeBuilder()
			rpeBuilder2.addAxis(parent)
			rpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
			rpeBuilder2.addAxis(ancestorOrSelf)
			rpeBuilder2.addUdpeTest(newElementTest("*", nil, nil)) // <- actual tested match
			rpeBuilder2.addAxis(parent)
			rpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				rpeBuilder                           rpeBuilder
				expectedRpePathPatternReprAfterMatch string
			}{
				{
					rpeBuilder:                           rpeBuilder1,
					expectedRpePathPatternReprAfterMatch: `a\ε`,
				},
				{
					rpeBuilder:                           rpeBuilder2,
					expectedRpePathPatternReprAfterMatch: `\\a\ε`,
				},
			}

			for _, test := range tests {
				rpe := test.rpeBuilder.end()
				rpePathPattern := rpe.entryPoint()
				rpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)

				t.Run(`rpe path pattern=`+rpePathPattern.String(), func(t *testing.T) {
					_, newPathPattern, ok := rpePathPattern.matchWithReductionOf(positiveReduction, true)

					if !ok {
						t.Error(`rpe path pattern does NOT match a compatible Reduction`)
					}

					if newPathPattern != nil {
						t.Error(`rpe path pattern returns a new path pattern when it should NOT`)
					}
				})
			}
		})
	})
}
