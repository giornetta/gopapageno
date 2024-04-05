package xpath

import (
	"container/list"
	"fmt"
)

type speculation interface {
	evaluate(v evaluator) customBool
}

type speculationImpl struct {
	evaluationsCount int
	prd              predicate
	ctx              NonTerminal
	el               *list.Element
}

func (sp *speculationImpl) String() string {
	return fmt.Sprintf("(%v , %v)", sp.ctx, sp.prd)
}

type evaluator func(id int, context NonTerminal, evaluationsCount int) customBool

func (sp *speculationImpl) evaluate(v evaluator) (result customBool) {
	defer func() {
		sp.evaluationsCount++
	}()

	result = Undefined
	predicateAtomsIDs := sp.prd.atomsIDs()
	for _, atomID := range predicateAtomsIDs {
		id := int(atomID)
		atomValue := v(id, sp.ctx, sp.evaluationsCount)
		result = sp.prd.earlyEvaluate(atomID, atomValue)
		if result != Undefined {
			return result
		}
	}
	return
}

type speculationList interface {
	addSpeculation(prd predicate, ctx NonTerminal) speculation
	removeSpeculation(sp speculation) (ok bool)
	newIterator() speculationListIterator
	iterate(callback speculationListIterableCallback)
	len() int
}

type speculationListImpl struct {
	list *list.List
}

func newSpeculationList() speculationList {
	return &speculationListImpl{
		list: list.New(),
	}
}

func (spl *speculationListImpl) addSpeculation(prd predicate, ctx NonTerminal) speculation {
	sp := &speculationImpl{
		prd: prd,
		ctx: ctx,
	}
	sp.el = spl.list.PushBack(sp)
	return sp
}

func (spl *speculationListImpl) removeSpeculation(sp speculation) (ok bool) {
	spImpl, ok := sp.(*speculationImpl)
	if ok {
		spl.list.Remove(spImpl.el)
		spImpl.prd = nil //avoid memory leak
		spImpl.ctx = nil //avoid memory leak
		spImpl.el = nil  //avoid memory leak
	}
	return
}

func (spl *speculationListImpl) len() int {
	return spl.list.Len()
}

// iterator object
type speculationListIterator interface {
	next() (sp speculation, hasNext bool)
	hasNext() bool
}

type speculationListIteratorImpl struct {
	nextEl *list.Element
}

func (splIt *speculationListIteratorImpl) next() (sp speculation, hasNext bool) {
	sp, ok := splIt.nextEl.Value.(speculation)
	if !ok {
		panic(`speculation list iterator error: trying to access a non existing next speculation`)
	}
	splIt.nextEl = splIt.nextEl.Next()
	hasNext = splIt.nextEl != nil
	return
}

func (splIt *speculationListIteratorImpl) hasNext() bool {
	return splIt.nextEl != nil
}

func (spl *speculationListImpl) newIterator() speculationListIterator {
	return &speculationListIteratorImpl{
		nextEl: spl.list.Front(),
	}
}

// iterate method
type speculationListIterableCallback func(sp speculation) (doBreak bool)

func (spl *speculationListImpl) iterate(callback speculationListIterableCallback) {
	var next *list.Element
	for e := spl.list.Front(); e != nil; e = next {
		next = e.Next()
		speculation, ok := e.Value.(speculation)

		if !ok {
			panic(`speculation list iterate: can NOT access to the next speculation`)
		}

		if doBreak := callback(speculation); doBreak {
			return
		}
	}
}
