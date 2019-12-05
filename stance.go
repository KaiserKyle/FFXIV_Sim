package main

import "strconv"

type stance interface {
	activate()
	reset()
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
	d.Name = "Blood Of The Dragon"
	d.Timer = 30.0
	d.Active = true
	d.Stacks = 0

	globalLog(Basic, "[STANCE ACTIVATED] "+d.Name)
}

func (d *dragoonStance) reset() {
	d.Timer = 0
	d.Active = false
	d.Stacks = 0
}

func (d *dragoonStance) isActive() bool {
	return d.Active
}

func (d *dragoonStance) advanceTime(span float64) {
	if d.Active {
		d.Timer -= span
	}

	if d.Active && d.Timer <= 0.0 {
		if d.Name == "Blood Of The Dragon" {
			d.Active = false
			globalLog(Basic, "[STANCE DEACTIVATED] "+d.Name)
		} else {
			d.Name = "Blood Of The Dragon"
			d.Timer = 30.0
			globalLog(Basic, "[STANCE DEMOTED] "+d.Name)
		}
	}
}

func (d *dragoonStance) checkSkill(skillName string) bool {
	if skillName == "Blood Of The Dragon" {
		if d.Active && d.Name == "Life Of The Dragon" {
			return false
		}
		return true
	}
	if d.Active {
		if skillName == "Nastrond" && d.Name != "Life Of The Dragon" {
			return false
		}
		if skillName == "Stardiver" && d.Name != "Life Of The Dragon" {
			return false
		}
		return true
	}
	return false
}

func (d *dragoonStance) checkPotencyBoost(skillName string) float64 {
	if d.Active {
		if skillName == "High Jump" || skillName == "Spineshatter Dive" {
			return 1.30
		}
	}
	return 1.00
}

// Returns applied effect
func (d *dragoonStance) processSkillExecution(result *skillResult) string {
	if d.Active {
		if result.SkillName == "Full Thrust" && result.IsCombo {
			return "Sharper Fang And Claw"
		} else if result.SkillName == "Chaos Thrust" && result.IsCombo {
			return "Enhanced Wheeling Thrust"
		} else if result.SkillName == "Fang And Claw" && !result.IsCombo {
			return "Enhanced Wheeling Thrust"
		} else if result.SkillName == "Wheeling Thrust" && !result.IsCombo {
			return "Sharper Fang And Claw"
		}
	}

	return ""
}

// Skills that cause a change in the stance will alert the stance when they
// are executed, allowing the stance to execute special logic
func (d *dragoonStance) alertSkill(skillName string) {
	if d.Active && d.Name == "Blood Of The Dragon" {
		if skillName == "Fang And Claw" || skillName == "Wheeling Thrust" || skillName == "Sonic Thrust" || skillName == "Coerthan Torment" {
			d.Timer += 10.0
			if d.Timer > 30.0 {
				d.Timer = 30.0
			}
			globalLogFloat(Basic, "[STANCE TIMER INCREASED] ", d.Timer)
		} else if skillName == "Geirskogul" && d.Stacks == 2 {
			d.Name = "Life Of The Dragon"
			d.Stacks = 0
			d.Timer = 30.0
			globalLog(Basic, "[STANCE PROMOTED] "+d.Name)
		}
	}
	if d.Active && skillName == "Mirage Dive" {
		d.Stacks++
		if d.Stacks > 2 {
			d.Stacks = 2
		}
		globalLog(Basic, "[STANCE STACKS INCREASED] "+strconv.Itoa(d.Stacks))
	}
}
