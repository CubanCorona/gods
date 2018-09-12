package nfa

import (
	"github.com/paulgriffiths/gods/automata/dfa"
	"github.com/paulgriffiths/gods/sets"
)

// Nfa implements a nondeterministic finite automaton.
type Nfa struct {
	Q      int                    // Number of states
	S      []rune                 // Alphabet
	D      []map[rune]sets.SetInt // Transition function
	Start  int                    // Start state
	Accept sets.SetInt            // Set of accepting states
}

// Accepts returns true if the NFA accepts the provided string.
func (n Nfa) Accepts(input string) bool {
	current := n.Eclosure(sets.NewSetInt(n.Start))

	for _, letter := range input {
		next := sets.NewSetInt()
		for _, state := range current.Elements() {
			if p, ok := n.D[state][letter]; ok {
				next = next.Union(n.Eclosure(p))
			}
		}
		if current.IsEmpty() {
			return false
		}
		current = next
	}

	return !n.Accept.Intersection(current).IsEmpty()
}

// Eclosure returns the set of states reachable from the provided
// set of states on e-transitions alone, including the provided
// set of states itself.
func (n Nfa) Eclosure(s sets.SetInt) sets.SetInt {
	current := s
	ec := s
	prevLen := -1

	for ec.Length() != prevLen {
		prevLen = ec.Length()
		next := sets.NewSetInt()
		for _, state := range current.Elements() {
			if eStates, ok := n.D[state][0]; ok {
				ec = ec.Union(eStates)
				next = next.Union(eStates)
			}
		}
		current = next
	}

	return ec
}

// ToDfa returns an equivalent Dfa
//
// Sources:
//	https://www.cs.odu.edu/~toida/nerzic/390teched/regular/fa/nfa-2-dfa.html
//	http://web.cecs.pdx.edu/~harry/compilers/slides/LexicalPart3.pdf
func (n Nfa) ToDfa() (dfa.Dfa, error) {
	var d dfa.Dfa
	// Track old states corresponding to each new state
	var tracker map[int]sets.SetInt = make(map[int]sets.SetInt)

	// Initialize Dfa and compute e-closure of starting state
	d.Q = 1
	d.S = n.S // Alphabet
	d.Start = 0
	d.Accept = sets.NewSetInt()

	tracker[0] = n.Eclosure(sets.NewSetInt(n.Start))

	// For each new state in the Dfa
	for i := 0; i < d.Q; i++ {

		// Make transition map
		d.D = append(d.D, make(map[rune]int))
		// Make transition tracker
		var transitions map[rune]sets.SetInt = make(map[rune]sets.SetInt)

		// For each old state in Nfa (that formed this new state in Dfa)
		for _, oldState := range tracker[i].Elements() {

			// For each transition of each old state
			for s, _ := range n.D[oldState] {

				// Add these transitions (and e-closures thereof)
				if s == 0 { // Ignore if this is an e-transition (will be captured by e-closures)
					continue
				}
				_, exists := transitions[s]
				if !exists {
					transitions[s] = sets.NewSetInt()
				}

				transitions[s] = transitions[s].Union(n.Eclosure(n.D[oldState][s]))
			}
		}

		// Make new states
	TransitionLoop:
		for s, t := range transitions {

			// Check existing (new) Dfa states
			for newState, oldStates := range tracker {
				if oldStates.Equals(t) {
					// Add transition to existing (new) Dfa state

					d.D[i][s] = newState
					continue TransitionLoop
				}
			}
			// No corresponding Dfa state exists -- make new Dfa state
			d.D[i][s] = d.Q
			tracker[d.Q] = t
			// Check if new Dfa state is accepting
			if !t.Intersection(n.Accept).IsEmpty() {
				d.Accept.Insert(d.Q)
			}
			// Incremement the state count
			d.Q = d.Q + 1

		}
	}

	return d, nil

}
