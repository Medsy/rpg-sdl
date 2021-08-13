package game

import (
	"fmt"
)

// Attack c1 attacks c2
func Attack(c1, c2 *Character) []string {
	var events []string
	c2.Hitpoints -= c1.Strength
	c1.AP--
	events = append(events, fmt.Sprintf("%s attacked %s for %d damage", c1.Name, c2.Name, c1.Strength))

	return events
}
