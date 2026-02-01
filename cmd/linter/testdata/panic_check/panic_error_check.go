package panic_check

func errBaseCheck() {
	panic("panic")
}

func errCheckInFunc() {
	func() {
		panic("panic")
	}()
}

func errCheckInIf() {
	if true {
		panic("error")
	}
}

func errCheckInFor() {
	for i := 0; i < 10; i++ {
		panic("error")
	}
}

func errCheckInSwitch() {
	switch {
	case true:
		panic("error")
	}
}

func errCheckWithVariable() {
	err := "error message"
	panic(err)
}

func errCheckInDefer() {
	defer func() {
		panic("error")
	}()
}

func errCheckInGoroutine() {
	go func() {
		panic("error")
	}()
}

func errCheckWithNil() {
	panic(nil)
}

// MyStruct Тестовая структура
type MyStruct struct{}

func (m MyStruct) errCheckInMethod() {
	panic("error")
}

func (m *MyStruct) errCheckInPointerMethod() {
	panic("error")
}

func errCheckInNestedBlocks() {
	{
		{
			panic("error")
		}
	}
}

func errCheckWithErrorType() {
	var err error
	panic(err)
}

func errCheckInSelect() {
	select {
	default:
		panic("error")
	}
}
