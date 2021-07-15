package main

import (
	"fmt"
	"strings"
	"./classpath"
	"./classfile"

)
import "./instructions"
import "./instructions/base"
import "./rtda"

type Cmd struct {
	helpFlag    bool
	versionFlag bool
	cpOption    string
	XjreOption  string
	class       string
	args        []string
}

func main() {
	cmd := &Cmd{}
	cmd.XjreOption = "/Library/Java/JavaVirtualMachines/jdk1.8.0_144.jdk/Contents/Home/jre"
	cmd.cpOption = "/Users/luotao/IdeaProjects/tluo-jvm/target/com.myjvm-1.1-SNAPSHOT.jar"
	cmd.class = "com.test.GaussTest"

	fmt.Printf("classpath:%s class:%s args:%v\n",
		cmd.cpOption, cmd.class, cmd.args)


	startJVM(cmd)
}

func startJVM(cmd *Cmd){
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)

	className := strings.Replace(cmd.class, ".", "/", -1)

	cf := loadClass(className, cp)
	mainMethod := getMainMethod(cf)
	fmt.Println(mainMethod.Name())
	fmt.Println(mainMethod.Descriptor())
	if mainMethod != nil {
		interpret(mainMethod)
	} else {
		fmt.Printf("Main method not found in class %s\n", cmd.class)
	}
}
func loadClass(className string, cp *classpath.Classpath) *classfile.ClassFile {
	classData, _, err := cp.ReadClass(className)
	if err != nil {
		panic(err)
	}

	cf, err := classfile.Parse(classData)
	if err != nil {
		panic(err)
	}

	return cf
}
func getMainMethod(cf *classfile.ClassFile) *classfile.MemberInfo {
	for _, m := range cf.Methods() {
		if m.Name() == "main" && m.Descriptor() == "([Ljava/lang/String;)V" {
			return m
		}
	}
	return nil
}


func interpret(methodInfo *classfile.MemberInfo) {
	codeAttr := methodInfo.CodeAttribute()
	maxLocals := codeAttr.MaxLocals()
	maxStack := codeAttr.MaxStack()
	bytecode := codeAttr.Code()

	thread := rtda.NewThread()
	frame := thread.NewFrame(maxLocals, maxStack)
	thread.PushFrame(frame)

	defer catchErr(frame)
	loop(thread, bytecode)
}

func catchErr(frame *rtda.Frame) {
	if r := recover(); r != nil {
		fmt.Printf("LocalVars:%v\n", frame.LocalVars())
		fmt.Printf("OperandStack:%v\n", frame.OperandStack())
		//panic(r)
	}
}


func loop(thread *rtda.Thread, bytecode []byte) {
	frame := thread.PopFrame()
	reader := &base.BytecodeReader{}

	for {
		pc := frame.NextPC()
		thread.SetPC(pc)

		// decode
		reader.Reset(bytecode, pc)
		opcode := reader.ReadUint8()
		inst := instructions.NewInstruction(opcode)
		inst.FetchOperands(reader)
		frame.SetNextPC(reader.PC())

		// execute
		fmt.Printf("pc:%2d inst:%T %v\n", pc, inst, inst)
		inst.Execute(frame)
	}
}