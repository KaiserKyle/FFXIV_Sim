package main

import "fmt"

const timeIncrement float64 = 0.01

type simulator struct {
	Players    []*playerCharacter
	Enemies    []*enemy
	SkillQueue []string
}

func (s *simulator) Run() {
	currentSkillIndex := 0
	critCount := 0
	directCount := 0
	critDirectCount := 0
	totalDamageDone := 0
	timeEllapsed := 0.0

	for currentSkillIndex < len(s.SkillQueue) {
		for i := range s.Players {
			enemies[0].advanceTime(timeIncrement)
			result := s.Players[i].advanceTime(timeIncrement, enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalDamageDone += result.DamageDone
				if result.DidCrit {
					critCount++
				}
				if result.DidDirect {
					directCount++
				}
				if result.DidCrit && result.DidDirect {
					critDirectCount++
				}
			}
			timeEllapsed += timeIncrement
			result = s.Players[i].performAction(s.SkillQueue[currentSkillIndex], enemies[0])
			if result != nil {
				enemies[0].applyDamage(result)
				result.Log()
				result.ApplyBuffs(*s.Players[i])
				totalDamageDone += result.DamageDone
				currentSkillIndex++
				if result.DidCrit {
					critCount++
				}
				if result.DidDirect {
					directCount++
				}
				if result.DidCrit && result.DidDirect {
					critDirectCount++
				}
			}
		}
	}

	fmt.Printf("Total Damage Done: %d\n", totalDamageDone)
	fmt.Printf("Time ellapsed: %f\n", timeEllapsed)
	fmt.Printf("DPS: %f\n", float64(totalDamageDone)/timeEllapsed)
	fmt.Printf("Skills performed: %d\n", currentSkillIndex)
	fmt.Printf("Crit Rate: %.2f\n", float64(critCount)/float64(currentSkillIndex))
	fmt.Printf("Direct Rate: %.2f\n", float64(directCount)/float64(currentSkillIndex))
	fmt.Printf("CritDirect Rate: %.2f\n", float64(critDirectCount)/float64(currentSkillIndex))
}
