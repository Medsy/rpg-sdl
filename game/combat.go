package game

import (
	"fmt"
)

// Attack c1 attacks c2
func Attack(c1, c2 *Character) {
	c2.Hitpoints -= c1.Strength
	c1.AP--
	fmt.Printf("%s attacked %s for %d\n", c1.Name, c2.Name, c1.Strength)

	if c2.Hitpoints > 0 {
		c1.Hitpoints -= c2.Strength
		fmt.Printf("%s attacked %s back for %d\n", c2.Name, c1.Name, c2.Strength)
		fmt.Printf("%s: %dhp\n %s: %dhp\n", c1.Name, c1.Hitpoints, c2.Name, c2.Hitpoints)
	} else {
		fmt.Printf("%s died\n", c2.Name)
	}

}
