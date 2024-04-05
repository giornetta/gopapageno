package xpath

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFpeInnerTest(t *testing.T) {
	t.Run(`fpeInnerTest.String()`, func(t *testing.T) {
		t.Run(`fpeInnerTest.behindDescendantAxis==true`, func(t *testing.T) {
			elementTest := newElementTest("a", nil, nil)
			fpeInnerTest := &fpeInnerTestImpl{
				udpeTest:             elementTest,
				behindDescendantAxis: true,
			}

			want := fmt.Sprintf("%v//", elementTest)
			if got := fpeInnerTest.String(); got != want {
				t.Errorf(`fpeInnerTest.String()=%v | want %v`, got, want)
			}
		})

		t.Run(`fpeInnerTest.behindDescendantAxis==false`, func(t *testing.T) {
			elementTest := newElementTest("a", nil, nil)
			fpeInnerTest := &fpeInnerTestImpl{
				udpeTest:             elementTest,
				behindDescendantAxis: false,
			}

			t.Run(`fpeInnerTest.isEntry==false`, func(t *testing.T) {
				fpeInnerTest.isEntry = false

				want := fmt.Sprintf("%v/", elementTest)
				if got := fpeInnerTest.String(); got != want {
					t.Errorf(`fpeInnerTest.String()=%v | want %v`, got, want)
				}
			})

			t.Run(`fpeInnerTest.isEntry==true`, func(t *testing.T) {
				fpeInnerTest.isEntry = true

				want := fmt.Sprintf("%v", elementTest)
				if got := fpeInnerTest.String(); got != want {
					t.Errorf(`fpeInnerTest.String()=%v | want %v`, got, want)
				}
			})

		})
	})

	t.Run(`fpeInnerTest.matchWithReductionOf(n)`, func(t *testing.T) {
		t.Run(`fpeInnerTest.matchWithReductionOf(n)=_,_,_,_,false if udpeTest does NOT match`, func(t *testing.T) {
			fpeInnerTest := &fpeInnerTestImpl{
				udpeTest: newElementTest("a", nil, nil),
			}

			if _, _, _, _, ok := fpeInnerTest.matchWithReductionOf(newElement("b", nil, nil)); ok {
				t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,_,_,_,true | want _,_,_,_,false`)
			}
		})

		t.Run(`fpeInnerTest.matchWithReductionOf(n)=_,next,newTest,true if udpeTest does match`, func(t *testing.T) {
			fpeInnerTest := &fpeInnerTestImpl{
				udpeTest: newElementTest("a", nil, nil),
			}

			if _, _, _, _, ok := fpeInnerTest.matchWithReductionOf(newElement("a", nil, nil)); !ok {
				t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,_,_,_,false | want _,_,_,_,true`)
			}

			t.Run(`when behindDescendantAxis==false`, func(t *testing.T) {
				precedingFpeInnerTest := &fpeInnerTestImpl{}
				fpeInnerTest := &fpeInnerTestImpl{
					udpeTest:              newElementTest("a", nil, nil),
					precedingFpeInnerTest: precedingFpeInnerTest,
				}

				_, next, newTest, hasNewTest, _ := fpeInnerTest.matchWithReductionOf(newElement("a", nil, nil))

				if next != precedingFpeInnerTest {
					t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,next,_,_ returns wrong next`)
				}

				if newTest != nil || hasNewTest {
					t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,_,newTest,hasNewTest,_ returns wrongly`)
				}

			})

			t.Run(`when behindDescendantAxis==true`, func(t *testing.T) {
				t.Run(`when isEntry==false && precedingFpeInnerTest != nil && precedingFpeInnerTest.behindDescendantAxis==true`, func(t *testing.T) {
					precedingFpeInnerTest := &fpeInnerTestImpl{
						behindDescendantAxis: true,
					}
					fpeInnerTest := &fpeInnerTestImpl{
						isEntry:               false,
						udpeTest:              newElementTest("a", nil, nil),
						precedingFpeInnerTest: precedingFpeInnerTest,
						behindDescendantAxis:  true,
					}

					_, next, newTest, hasNewTest, _ := fpeInnerTest.matchWithReductionOf(newElement("a", nil, nil))

					if next != precedingFpeInnerTest {
						t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,next,_,_,_ returns wrong next`)
					}

					if newTest != nil || hasNewTest {
						t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,_,newTest,hasNewTest,_ returns wrongly`)
					}
				})

				t.Run(`otherwise, when isEntry==true`, func(t *testing.T) {
					precedingFpeInnerTest := &fpeInnerTestImpl{
						behindDescendantAxis: true,
					}
					fpeInnerTest := &fpeInnerTestImpl{
						isEntry:               true,
						udpeTest:              newElementTest("a", nil, nil),
						precedingFpeInnerTest: precedingFpeInnerTest,
						behindDescendantAxis:  true,
					}

					_, next, newTest, hasNewTest, _ := fpeInnerTest.matchWithReductionOf(newElement("a", nil, nil))

					if next != fpeInnerTest {
						t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,next,_,_,_ returns wrong next`)
					}

					if newTest != precedingFpeInnerTest || !hasNewTest {
						t.Error(`fpeInnerTest.matchWithReductionOf(n)=_,_,newTest,hasNewTest,_ returns wrongly`)
					}
				})
			})
		})
	})
}
func TestFpeBuilder(t *testing.T) {
	t.Run(`fpeBuilder.addUdpeTest(t)`, func(t *testing.T) {
		t.Run(`fpeBuilder.addUdpeTest(t)=false if fpeBuilder expecting an axis`, func(t *testing.T) {
			fpeBuilder := new(fpeBuilderImpl)

			if fpeBuilder.addUdpeTest(newElementTest("a", nil, nil)) {
				t.Error(`fpeBuilder.addUdpeTest(t)=true | want false`)
			}
		})

		t.Run(`fpeBuilder.addUdpeTest(t)=true if fpeBuilder expecting an udpeTest`, func(t *testing.T) {
			fpeBuilder := new(fpeBuilderImpl)
			fpeBuilder.state = expectFpeUdpeTest

			udpeTest := newElementTest("a", nil, nil)
			if !fpeBuilder.addUdpeTest(udpeTest) {
				t.Error(`fpeBuilder.addUdpeTest(t)=false | want true`)
			}

			if fpeBuilder.state != expectFpeAxis {
				t.Error(`fpeBuilder.addUdpeTest(t) doesn't update fpeBuilder.state correctly`)
			}

			expectedNewPrecedentFpeInnerTest := &fpeInnerTestImpl{
				udpeTest:              udpeTest,
				precedingFpeInnerTest: nil,
			}

			if got := fpeBuilder.precedentFpeInnerTest; !reflect.DeepEqual(*got, *expectedNewPrecedentFpeInnerTest) {
				t.Error(`fpeBuilder.addUdpeTest(t) doesn't update fpeBuilder.precedentFpeInnerTest correctly`)
			}
		})
	})

	t.Run(`fpeBuilder.addAxis(a)`, func(t *testing.T) {
		t.Run(`fpeBuilder.addAxis(a)=false if fpeBuilder expecting an udpeTest`, func(t *testing.T) {
			fpeBuilder := new(fpeBuilderImpl)
			fpeBuilder.state = expectFpeUdpeTest

			if fpeBuilder.addAxis(child) {
				t.Error(`fpeBuilder.addAxis(a)=true | want false`)
			}
		})

		t.Run(`fpeBuilder.addAxis(a)=true if fpeBuilder expecting an axis`, func(t *testing.T) {
			t.Run(`fpeBuilder.precedentFpeInnerTest=nil`, func(t *testing.T) {
				t.Run(`a=child`, func(t *testing.T) {
					fpeBuilder := new(fpeBuilderImpl)

					if !fpeBuilder.addAxis(child) {
						t.Error(`fpeBuilder.addAxis(child)=false | want true`)
					}

					if fpeBuilder.state != expectFpeUdpeTest {
						t.Error(`fpeBuilder.addAxis(child) doesn't update fpeBuilder.state correctly`)
					}
				})

				t.Run(`a=descendantOrSelf`, func(t *testing.T) {
					fpeBuilder := new(fpeBuilderImpl)

					if !fpeBuilder.addAxis(descendantOrSelf) {
						t.Error(`fpeBuilder.addAxis(descendantOrSelf)=false | want true`)
					}

					if fpeBuilder.state != expectFpeUdpeTest {
						t.Error(`fpeBuilder.addAxis(descendantOrSelf) doesn't update fpeBuilder.state correctly`)
					}
				})
			})

			t.Run(`fpeBuilder.precedentFpeInnerTest !=nil`, func(t *testing.T) {
				t.Run(`a=child`, func(t *testing.T) {
					fpeBuilder := new(fpeBuilderImpl)
					fpeInnerTest := new(fpeInnerTestImpl)
					fpeBuilder.precedentFpeInnerTest = new(fpeInnerTestImpl)

					if !fpeBuilder.addAxis(child) {
						t.Error(`fpeBuilder.addAxis(child)=false | want true`)
					}

					if fpeBuilder.state != expectFpeUdpeTest {
						t.Error(`fpeBuilder.addAxis(child) doesn't update fpeBuilder.state correctly`)
					}

					if !reflect.DeepEqual(*fpeBuilder.precedentFpeInnerTest, *fpeInnerTest) {
						t.Error(`fpeBuilder.addAxis(child) shouldn't update fpeBuilder.precedentFpeInnerTest`)
					}
				})

				t.Run(`a=descendantOrSelf`, func(t *testing.T) {
					fpeBuilder := new(fpeBuilderImpl)
					fpeBuilder.precedentFpeInnerTest = new(fpeInnerTestImpl)

					if !fpeBuilder.addAxis(descendantOrSelf) {
						t.Error(`fpeBuilder.addAxis(descendantOrSelf)=false | want true`)
					}

					if fpeBuilder.state != expectFpeUdpeTest {
						t.Error(`fpeBuilder.addAxis(descendantOrSelf) doesn't update fpeBuilder.state correctly`)
					}

					if precedentFpeInnerTest := fpeBuilder.precedentFpeInnerTest; !precedentFpeInnerTest.behindDescendantAxis {
						t.Error(`fpeBuilder.addAxis(descendantOrSelf) doesn't update fpeBuilder.precedentFpeInnerTest correctly`)
					}

				})
			})
		})
	})

	t.Run(`fpeBuilder.end()`, func(t *testing.T) {
		t.Run(`fpeBuilder.end()=nil if no udpeTest has been added`, func(t *testing.T) {
			fpeBuilder := newFpeBuilder()

			if fpeBuilder.end() != nil {
				t.Error(`fpeBuilder.end()!=nil | want nil`)
			}
		})

		t.Run(`fpeBuilder.end() returns a fpe if at least one udpeTest has been added`, func(t *testing.T) {
			fpeBuilder := newFpeBuilder()
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))

			fpe := fpeBuilder.end().(*fpeImpl)
			if fpe == nil {
				t.Error(`fpeBuilder.end()=nil | want fpe`)
			}

			if !fpe.entryTest.isEntry {
				t.Error(`fpeBuilder.end().entryTest.isEntry=false | want true`)
			}
		})
	})
}
func TestFpe(t *testing.T) {
	t.Run(`fpe.entryPoint()`, func(t *testing.T) {
		fpeBuilder := newFpeBuilder()
		fpeBuilder.addAxis(child)
		fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
		fpe := fpeBuilder.end().(*fpeImpl)

		entryPoint := fpe.entryPoint().(*fpePathPatternImpl)
		if entryPoint.currentTest != fpe.entryTest {
			t.Error(`fpe.entryPoint().currentTest != fpe.entryTest`)
		}

	})

	t.Run(`fpe.String()`, func(t *testing.T) {
		t.Run(`fpe=a/b`, func(t *testing.T) {
			fpeBuilder := newFpeBuilder()
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			fpe := fpeBuilder.end()
			want := "a/b"
			if got := fpe.String(); got != want {
				t.Errorf(`fpe.String()=%v | want %v`, got, want)
			}
		})

		t.Run(`fpe=a//b`, func(t *testing.T) {
			fpeBuilder := newFpeBuilder()
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder.addAxis(descendantOrSelf)
			fpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			fpe := fpeBuilder.end()

			want := "a//b"
			if got := fpe.String(); got != want {
				t.Errorf(`fpe.String()=%v | want %v`, got, want)
			}
		})
	})
}

// Test integration that makes up Algorithm 1
func TestFpeIntegration(t *testing.T) {
	t.Run(`Phase 1`, func(t *testing.T) {
		t.Run(`1. γ = α<b>X</b>`, func(t *testing.T) {
			reducedElement := newElement("b", []*Attribute{NewAttribute("key", "value")}, nil)

			t.Run(`2. 4. 11.`, func(t *testing.T) {
				//2.
				fpeBuilder1 := newFpeBuilder()
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))

				fpeBuilder2 := newFpeBuilder()
				fpeBuilder2.addAxis(child)
				fpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder2.addAxis(child)
				fpeBuilder2.addUdpeTest(newElementTest("*", nil, nil))
				//4.
				fpeBuilder3 := newFpeBuilder()
				fpeBuilder3.addAxis(child)
				fpeBuilder3.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder3.addAxis(descendantOrSelf)
				fpeBuilder3.addUdpeTest(newElementTest("b", nil, nil))

				fpeBuilder4 := newFpeBuilder()
				fpeBuilder4.addAxis(child)
				fpeBuilder4.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder4.addAxis(descendantOrSelf)
				fpeBuilder4.addUdpeTest(newElementTest("*", nil, nil))
				//11.
				fpeBuilder5 := newFpeBuilder()
				fpeBuilder5.addAxis(child)
				fpeBuilder5.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder5.addAxis(child)
				fpeBuilder5.addUdpeTest(newElementTest("b", NewAttribute("key", "value"), nil))

				fpeBuilder6 := newFpeBuilder()
				fpeBuilder6.addAxis(child)
				fpeBuilder6.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder6.addAxis(descendantOrSelf)
				fpeBuilder6.addUdpeTest(newElementTest("b", NewAttribute("key", "value"), nil))

				const expectedPathPatternReprAfterMatch = "ε/a"
				var tests = []struct {
					fpeBuilder                        fpeBuilder
					expectedPathPatternReprAfterMatch string
				}{
					{
						fpeBuilder:                        fpeBuilder1,
						expectedPathPatternReprAfterMatch: "ε/a",
					},
					{
						fpeBuilder:                        fpeBuilder2,
						expectedPathPatternReprAfterMatch: "ε/a",
					},
					{
						fpeBuilder:                        fpeBuilder3,
						expectedPathPatternReprAfterMatch: "ε/a//",
					},
					{
						fpeBuilder:                        fpeBuilder4,
						expectedPathPatternReprAfterMatch: "ε/a//",
					},
					{
						fpeBuilder:                        fpeBuilder5,
						expectedPathPatternReprAfterMatch: "ε/a",
					},
					{
						fpeBuilder:                        fpeBuilder6,
						expectedPathPatternReprAfterMatch: "ε/a//",
					},
				}

				for _, test := range tests {
					fpe := test.fpeBuilder.end()
					t.Run(`fpe=`+fpe.String(), func(t *testing.T) {
						entryPoint := fpe.entryPoint()

						_, newPathPattern, ok := entryPoint.matchWithReductionOf(reducedElement, true)

						if !ok {
							t.Error(`fpe entry point does NOT match compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe entry point returns a new fpe path pattern when it should NOT`)
						}

						if entryPoint.String() != test.expectedPathPatternReprAfterMatch {
							t.Errorf(`fpe entry point does NOT update correcty`)
						}
					})
				}
			})

			t.Run(`6.`, func(t *testing.T) {
				fpeBuilder1 := newFpeBuilder()
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))
				fpeBuilder1.addAxis(descendantOrSelf)

				fpeBuilder2 := newFpeBuilder()
				fpeBuilder2.addAxis(child)
				fpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder2.addAxis(child)
				fpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))
				fpeBuilder2.addAxis(descendantOrSelf)

				fpeBuilder3 := newFpeBuilder()
				fpeBuilder3.addAxis(child)
				fpeBuilder3.addUdpeTest(newElementTest("b", nil, nil))
				fpeBuilder3.addAxis(descendantOrSelf)

				var tests = []struct {
					fpeBuilder            fpeBuilder
					returnsNewPathPattern bool
					newPathPatternRepr    string
				}{
					{
						fpeBuilder:            fpeBuilder1,
						returnsNewPathPattern: true,
						newPathPatternRepr:    "ε/a",
					},
					{
						fpeBuilder:            fpeBuilder2,
						returnsNewPathPattern: false,
					},
					{
						fpeBuilder:            fpeBuilder3,
						returnsNewPathPattern: true,
						newPathPatternRepr:    "ε",
					},
				}

				for _, test := range tests {
					fpe := test.fpeBuilder.end()
					t.Run(`fpe=`+fpe.String(), func(t *testing.T) {
						entryPoint := fpe.entryPoint()
						entryPointReprBeforeMatch := entryPoint.String()

						_, newPathPattern, ok := entryPoint.matchWithReductionOf(reducedElement, true)

						if !ok {
							t.Error(`fpe entry point does NOT match compatible Reduction`)
						}

						if entryPoint.String() != entryPointReprBeforeMatch {
							t.Error(`fpe entry point does NOT update correctly`)
						}

						if test.returnsNewPathPattern {
							if newPathPattern == nil {
								t.Error(`fpe entry point does NOT returns a new path pattern when it should`)
							} else if newPathPattern.String() != test.newPathPatternRepr {
								t.Error(`fpe entry point returns wrong new path pattern`)
							}

						} else {
							if newPathPattern != nil {
								t.Error(`fpe entry point returns a new path pattern when it should NOT`)
							}
						}
					})
				}
			})
		})

		t.Run(`13. γ = α STRING`, func(t *testing.T) {
			reducedText := newText("some Text", nil)

			t.Run(`14. 16.`, func(t *testing.T) {
				//16.
				fpeBuilder1 := newFpeBuilder()
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder1.addAxis(child)
				fpeBuilder1.addUdpeTest(newTextTest("some Text"))

				fpeBuilder2 := newFpeBuilder()
				fpeBuilder2.addAxis(child)
				fpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder2.addAxis(descendantOrSelf)
				fpeBuilder2.addUdpeTest(newTextTest("some Text"))
				//14.
				fpeBuilder3 := newFpeBuilder()
				fpeBuilder3.addAxis(child)
				fpeBuilder3.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder3.addAxis(child)
				fpeBuilder3.addUdpeTest(newTextTest(""))

				fpeBuilder4 := newFpeBuilder()
				fpeBuilder4.addAxis(child)
				fpeBuilder4.addUdpeTest(newElementTest("a", nil, nil))
				fpeBuilder4.addAxis(descendantOrSelf)
				fpeBuilder4.addUdpeTest(newTextTest(""))

				var tests = []struct {
					fpeBuilder                        fpeBuilder
					expectedPathPatternReprAfterMatch string
				}{
					{
						fpeBuilder:                        fpeBuilder1,
						expectedPathPatternReprAfterMatch: "ε/a",
					},
					{
						fpeBuilder:                        fpeBuilder2,
						expectedPathPatternReprAfterMatch: "ε/a//",
					},
					{
						fpeBuilder:                        fpeBuilder3,
						expectedPathPatternReprAfterMatch: "ε/a",
					},
					{
						fpeBuilder:                        fpeBuilder4,
						expectedPathPatternReprAfterMatch: "ε/a//",
					},
				}

				for _, test := range tests {
					fpe := test.fpeBuilder.end()
					t.Run(`fpe=`+fpe.String(), func(t *testing.T) {
						entryPoint := fpe.entryPoint()

						_, newPathPattern, ok := entryPoint.matchWithReductionOf(reducedText, true)

						if !ok {
							t.Error(`fpe entry point does NOT match compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe entry point returns a new fpe path pattern when it should NOT`)
						}

						if entryPoint.String() != test.expectedPathPatternReprAfterMatch {
							t.Errorf(`fpe entry point does NOT update correcty`)
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

		t.Run(`20.`, func(t *testing.T) {
			fpeBuilder1 := newFpeBuilder()
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("b", nil, nil)) // <- actual tested match
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			fpeBuilder2 := newFpeBuilder()
			fpeBuilder2.addAxis(child)
			fpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder2.addAxis(descendantOrSelf)
			fpeBuilder2.addUdpeTest(newElementTest("b", nil, nil)) // <- actual tested match
			fpeBuilder2.addAxis(child)
			fpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				fpeBuilder                        fpeBuilder
				expectedPathPatternReprAfterMatch string
			}{
				{
					fpeBuilder:                        fpeBuilder1,
					expectedPathPatternReprAfterMatch: "ε/a",
				},
				{
					fpeBuilder:                        fpeBuilder2,
					expectedPathPatternReprAfterMatch: "ε/a//",
				},
			}

			for _, test := range tests {
				fpe := test.fpeBuilder.end()
				fpePathPattern := fpe.entryPoint()
				fpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)

				t.Run(`fpe path pattern=`+fpePathPattern.String(), func(t *testing.T) {
					t.Run(`negative matching`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(negativeReduction, true)

						if ok {
							t.Error(`fpe path pattern matches an incompatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
						}
					})

					t.Run(`positive matching`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(positiveReduction, true)

						if !ok {
							t.Error(`fpe path pattern does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
						}

						if fpePathPattern.String() != test.expectedPathPatternReprAfterMatch {
							t.Error(`fpe path pattern does NOT update correctly`)
						}
					})
				})
			}

		})

		t.Run(`23.`, func(t *testing.T) {

			fpeBuilder1 := newFpeBuilder()
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("b", nil, nil))
			fpeBuilder1.addAxis(descendantOrSelf) //<- actual tested match
			fpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			fpeBuilder2 := newFpeBuilder()
			fpeBuilder2.addAxis(child)
			fpeBuilder2.addUdpeTest(newElementTest("b", nil, nil))
			fpeBuilder2.addAxis(descendantOrSelf) //<- actual tested match
			fpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				fpeBuilder                        fpeBuilder
				expectedPathPatternReprAfterMatch string
				newPathPatternRepr                string
			}{
				{
					fpeBuilder:                        fpeBuilder1,
					expectedPathPatternReprAfterMatch: "ε/a/b//",
					newPathPatternRepr:                "ε/a",
				},
				{
					fpeBuilder:                        fpeBuilder2,
					expectedPathPatternReprAfterMatch: "ε/b//",
					newPathPatternRepr:                "ε",
				},
			}

			for _, test := range tests {
				fpe := test.fpeBuilder.end()
				fpePathPattern := fpe.entryPoint()
				fpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)
				t.Run(`fpe path pattern=`+fpePathPattern.String(), func(t *testing.T) {
					t.Run(`25.`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(negativeReduction, true)

						if !ok {
							t.Error(`fpe path pattern does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
						}

						if fpePathPattern.String() != test.expectedPathPatternReprAfterMatch {
							t.Error(`fpe path pattern does NOT update correctly after match`)
						}
					})

					t.Run(`24.`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(positiveReduction, true)

						if !ok {
							t.Error(`fpe path pattern does NOT match a compatible Reduction`)
						}

						if fpePathPattern.String() != test.expectedPathPatternReprAfterMatch {
							t.Error(`fpe path pattern does NOT update correctly after match`)
						}

						if newPathPattern == nil || newPathPattern.String() != test.newPathPatternRepr {
							t.Error(`fpe path pattern returns wrong new path pattern`)
						}
					})
				})

			}

		})

		t.Run(`26.`, func(t *testing.T) {
			expectedPathPatternReprAfterMatch := "ε/a//"

			fpeBuilder := newFpeBuilder()
			fpeBuilder.addAxis(child)
			fpeBuilder.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder.addAxis(descendantOrSelf)
			fpeBuilder.addUdpeTest(newElementTest("b", nil, nil))
			fpeBuilder.addAxis(descendantOrSelf) //<- actual tested match
			fpeBuilder.addUdpeTest(newElementTest("c", nil, nil))

			fpe := fpeBuilder.end()
			fpePathPattern := fpe.entryPoint()
			fpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)
			fpePathPatternReprBeforeMatch := fpePathPattern.String()

			t.Run(`fpe path pattern=`+fpePathPattern.String(), func(t *testing.T) {
				t.Run(`28.`, func(t *testing.T) {
					_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(negativeReduction, true)

					if !ok {
						t.Error(`fpe path pattern does NOT match a compatible Reduction`)
					}

					if newPathPattern != nil {
						t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
					}

					if fpePathPattern.String() != fpePathPatternReprBeforeMatch {
						t.Error(`fpe path pattern does NOT update correctly`)
					}
				})

				t.Run(`27.`, func(t *testing.T) {
					_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(positiveReduction, true)

					if !ok {
						t.Error(`fpe path pattern does NOT match a compatible Reduction`)
					}

					if newPathPattern != nil {
						t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
					}

					if fpePathPattern.String() != expectedPathPatternReprAfterMatch {
						t.Error(`fpe path pattern does NOT update correctly`)
					}
				})
			})
		})

		t.Run(`29.`, func(t *testing.T) {
			fpeBuilder1 := newFpeBuilder()
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("*", nil, nil)) //<- actual tested match
			fpeBuilder1.addAxis(child)
			fpeBuilder1.addUdpeTest(newElementTest("c", nil, nil))

			fpeBuilder2 := newFpeBuilder()
			fpeBuilder2.addAxis(child)
			fpeBuilder2.addUdpeTest(newElementTest("a", nil, nil))
			fpeBuilder2.addAxis(descendantOrSelf)
			fpeBuilder2.addUdpeTest(newElementTest("*", nil, nil)) //<- actual tested match
			fpeBuilder2.addAxis(child)
			fpeBuilder2.addUdpeTest(newElementTest("c", nil, nil))

			var tests = []struct {
				fpeBuilder                        fpeBuilder
				expectedPathPatternReprAfterMatch string
			}{
				{
					fpeBuilder:                        fpeBuilder1,
					expectedPathPatternReprAfterMatch: "ε/a",
				},
				{
					fpeBuilder:                        fpeBuilder2,
					expectedPathPatternReprAfterMatch: "ε/a//",
				},
			}

			for _, test := range tests {
				fpe := test.fpeBuilder.end()
				fpePathPattern := fpe.entryPoint()
				fpePathPattern.matchWithReductionOf(entryPointConsumingReduction, true)
				t.Run(`fpe path pattern=`+fpePathPattern.String(), func(t *testing.T) {
					t.Run(`31.`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(newText("some Text", nil), true)

						if ok {
							t.Error(`fpe path pattern matches an incompatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
						}
					})

					t.Run(`29.`, func(t *testing.T) {
						_, newPathPattern, ok := fpePathPattern.matchWithReductionOf(positiveReduction, true)

						if !ok {
							t.Error(`fpe path pattern does NOT match a compatible Reduction`)
						}

						if newPathPattern != nil {
							t.Error(`fpe path pattern returns a new path pattern when it should NOT`)
						}

						if fpePathPattern.String() != test.expectedPathPatternReprAfterMatch {
							t.Error(`fpe path pattern does NOT update correctly`)
						}
					})
				})
			}
		})
	})
}
