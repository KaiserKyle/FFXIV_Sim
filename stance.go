package main

import "fmt"

type stance interface {
	activate()
	isActive() bool
	advanceTime(float64)
	processSkillExecution(*skillResult) string
	checkSkill(string) bool
	checkPotencyBoost(string) float64
	alertSkill(string)
}

type dragoonStance struct {
	Name   string
	Timer  float64
	Active bool
	Stacks int
}

func (d *dragoonStance) activate() {
	d.Name = "Blood of the Dragon"
	d.Timer = 20.0
	d.Active = true
	d.Stacks = 0

	fmt.Printf("[STANCE ACTIVATED] %s\n", d.Name)
}

func (d *dragoonStance) isActive() bool {
	return d.Active
}

func (d *dragoonStance) advanceTime(span float64) {
	if d.Active {
		d.Timer -= span
	}

	if d.Active && d.Timer <= 0.0 {
		if d.Name == "Blood of the Dragon" {
			d.Active = false
			fmt.Printf("[STANCE DEACTIVATED] %s\n", d.Name)
		} else {
			d.Name = "Blood of the Dragon"
			d.Timer = 20.0
			fmt.Printf("[STANCE DEMOTED] %s\n", d.Name)
		}
	}
}

func (d *dragoonStance) checkSkill(skillName string) bool {
	if skillName == "Blood of the Dragon" {
		if d.Active && d.Name == "Life of the Dragon" {
			return false
		}
		return true
	}
	if d.Active {
		if skillName == "Nastrond" && d.Name != "Life of the Dragon" {
			return false
		}
		return true
	}
	return false
}

func (d *dragoonStance) checkPotencyBoost(skillName string) float64 {
	if d.Active {
		if skillName == "Jump" || skillName == "Spineshatter Dive" {
			return 1.30
		}
	}
	return 1.00
}

// Returns applied effect
func (d *dragoonStance) processSkillExecution(result *skillResult) string {
	if d.Active {
		if result.SkillName == "Full Thrust" && result.IsCombo {
			return "Sharper Fang and Claw"
		} else if result.SkillName == "Chaos Thrust" && result.IsCombo {
			return "Enhanced Wheeling Thrust"
		} else if result.SkillName == "Fang and Claw" && !result.IsCombo {
			return "Enhanced Wheeling Thrust"
		} else if result.SkillName == "Wheeling Thrust" && !result.IsCombo {
			return "Sharper Fang and Claw"
		}
	}

	return ""
}

// Skills that cause a change in the stance will alert the stance when they
// are executed, allowing the stance to execute special logic
func (d *dragoonStance) alertSkill(skillName string) {
	if d.Active && d.Name == "Blood of the Dragon" {
		if skillName == "Fang and Claw" || skillName == "Wheeling Thrust" || skillName == "Sonic Thrust" {
			d.Timer += 10.0
			if d.Timer > 30.0 {
				d.Timer = 30.0
			}
			fmt.Printf("[STANCE TIMER INCREASED] %f\n", d.Timer)
		} else if skillName == "Geirskogul" && d.Stacks == 3 {
			d.Name = "Life of the Dragon"
			d.Stacks = 0
			if d.Timer < 20.0 {
				d.Timer = 20.0
			}
			fmt.Printf("[STANCE PROMOTED] %s\n", d.Name)
		}
	}
	if d.Active && skillName == "Mirage Dive" {
		d.Stacks++
		if d.Stacks > 3 {
			d.Stacks = 3
		}
		fmt.Printf("[STANCE STACKS INCREASED] %d\n", d.Stacks)
	}
}
