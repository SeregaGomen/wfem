package main

import (
	"fmt"
	"wfem/cmd/fem/fem"
	"wfem/cmd/fem/params"
	"wfem/cmd/web"
)

// 0. Переименовать интерфейсы
// 1. Проверка элемента (check_elm) при вычислении нагрузок должна быть разная для КЭ и ГЭ:
//   - для ГЭ проверять все узлы;
//   - для КЭ - только его центр
//
// 2. Длина отрезка (volume1d2) возможно неправилна реализована в fem (Rust)
// 3. Share не вычислять в цикле (QFEM и fem)
// 4. В fem (Rust) возможна ошибка при вычислении нормали к ГЭ в 2d (x или self.x?)
// 5. Parser для пустого выражения не использовать. Везде, где предикат пустой, обрабатывать эту ситуацию (QFEM и fem)
// 6. В QFEM доделать загрузку msh
// 7. Тоже для pyfem

func calcSquare() {
	meshName, resName := "data/square.mesh", "data/square.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(2)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddThickness("1", "")
	f.AddBoundaryCondition("0", "y == 0", params.X|params.Y)
	f.AddVolumeLoad("-0.05", "", params.Y)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcRail() {
	meshName, resName := "data/rail.mesh", "data/rail.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(4)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddBoundaryCondition("0", "y == 0", params.X|params.Y|params.Z)
	f.AddVolumeLoad("-100", "", params.Y)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcCube() {
	meshName, resName := "data/cube.mesh", "data/cube.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(4)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddBoundaryCondition("0", "z == 0", params.X|params.Y|params.Z)
	f.AddVolumeLoad("-0.5", "", params.Z)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcShelTube() {
	//meshName, resName := "data/shell3.mesh", "data/shell3.res"
	meshName, resName := "data/shell4.mesh", "data/shell4.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetEps(1.0e-10)
	f.SetNumThread(7)
	f.AddThickness("0.0369", "")
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddBoundaryCondition("0", "z == 0 or z == 4.014", params.X|params.Y|params.Z)
	f.AddPressureLoad("-0.05", "")
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcShel() {
	meshName, resName := "data/shell.mesh", "data/shell.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(1)
	f.SetEps(1.0e-10)
	f.AddThickness("0.0369", "")
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddBoundaryCondition("0", "y == -1 or y == 1", params.X|params.Y|params.Z)
	f.AddPressureLoad("-0.05", "")
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcConsole() {
	meshName, resName := "data/console.mesh", "data/console.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(4)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddThickness("1", "")
	f.AddBoundaryCondition("0", "x == 0", params.X|params.Y)
	//f.AddVolumeLoad("1", "", params.Y)
	//f.AddPointLoad("1", "x == 10 and y == -0.25", params.Y)
	f.AddSurfaceLoad("-1", "y == 0.25", params.Y)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcQuad() {
	meshName, resName := "data/quad.mesh", "data/quad.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(2)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddThickness("1", "")
	f.AddBoundaryCondition("0", "y == -0.5", params.X|params.Y)
	//f.AddVolumeLoad("500", "", params.Y)
	//f.AddPointLoad("1", "(x == -0.5 or x == 0.5) and y == 0.5", params.Y)
	f.AddSurfaceLoad("-0.05", "y == 0.5", params.Y)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcTank() {
	meshName, resName := "data/tank_1_4.mesh", "data/tank_1_4.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(4)
	f.AddVariable("C", 1.454)
	f.AddVariable("CX_BOT", 20.7657)
	f.AddVariable("CX_TOP", -8.5497)
	f.AddVariable("D", 3.9)
	f.AddVariable("FI_B", -0.872665)
	f.AddVariable("FI_T", -2.26893)
	f.AddVariable("H", 0.06)
	f.AddVariable("K2_BOT", 0.0520196)
	f.AddVariable("K2_TOP", 0.0520196)
	f.AddVariable("L", 12.216)
	f.AddVariable("L1", 1.767)
	f.AddVariable("L2", 2.122)
	f.AddVariable("L3", 1.654)
	f.AddVariable("L4", 1.09)
	f.AddVariable("P", 142196)
	f.AddVariable("R", 2.5)
	f.AddVariable("eps", 0.01)

	f.AddThickness("0.0046", "((abs(R * R - ((x - C) * (x - C) + y * y + z * z)) <= eps) and (x <= (R * cos(FI_T) + C))) or ((abs(R * R - ((x - L + C) * (x - L + C) + y * y + z * z)) <= eps) and (x >= (R * cos(FI_B) + L - C)))")
	f.AddThickness("0.05", "((x >= (R * cos(FI_T) + C)) and (x <= 0)) or ((x >= L) and (x <= (R * cos(FI_B) + L - C))) or ((x >= 4 * L3 - H / 2) and (x <= 4 * L3 + H / 2))")
	f.AddThickness("0.0255", "((x >= L3 - H / 2.0) and (x <= L3 + H / 2)) or ((x >= 2 * L3 - H / 2) and (x <= 2 * L3 + H / 2)) or ((x >= 5 * L3 - H / 2) and (x <= 5 * L3 + H / 2)) or ((x >= 6 * L3 - H / 2) and (x <= 6 * L3 + H / 2)) or ((x >= 6 * L3 - H / 2 + L4) and (x <= 6 * L3 + H / 2 + L4))")
	f.AddThickness("0.04", "(x >= 3 * L3 - H) and (x <= 3 * L3 + H)")
	f.AddThickness("0.0045", "(x >= 0 and x <= (L3 - H / 2)) or (x >= (L3 + H / 2) and x <= (2 * L3 - H / 2)) or (x >= (2 * L3 + H / 2) and x <= (3 * L3 - H)) or (x >= (4 * L3 + H / 2) and x <= (5 * L3 - H / 2)) or (x >= (5 * L3 + H / 2) and x <= (6 * L3 - H / 2))")
	f.AddThickness("0.0046", "x>= (3 * L3 + H) and x <= (4 * L3 - H / 2)")
	f.AddThickness("0.0052", "(x >= (6 * L3 + H / 2) and x <= (6 * L3 - H / 2 + L4)) or (x >= (6 * L3 + H / 2 + L4) and x <= L)")
	f.AddThickness("0.0143", "x < 0")
	f.AddThickness("0.016", "")

	f.AddYoungModulus("6.5e+10", "(abs(R * R - ((x - C) * (x - C) + y * y + z * z)) <= eps and x <= (R * cos(FI_T) + C)) or (abs(R * R - ((x - L + C) * (x - L + C) + y * y + z * z)) <= eps and x >= (R * cos(FI_B) + L - C))")
	f.AddYoungModulus("7.3e+10", "")
	f.AddPoissonRatio("0.3", "")

	f.AddPressureLoad("P", "x >= 0 and x <= L")
	f.AddPressureLoad("P", "abs(R * R - ((x - C) * (x - C) + y * y + z * z)) <= eps and x <= (R * cos(FI_T) + C)")
	f.AddPressureLoad("P", "abs(R * R - ((x - L + C) * (x - L + C) + y * y + z * z)) <= eps and x >= (R * cos(FI_B) + L - C)")
	f.AddPressureLoad("P", "(x>= (R * cos(FI_T) + C) and x <= 0) and abs(y ** 2 + z ** 2 - K2_TOP * (x - CX_TOP) ** 2) < eps")
	f.AddPressureLoad("P", "(x >= L and x <= (R * cos(FI_B) + L - C)) and abs(y ** 2 + z ** 2  - K2_BOT * (x - CX_BOT) ** 2) < eps")

	f.AddBoundaryCondition("0", "abs(x - 14.338) < eps", params.X|params.Y|params.Z)
	f.AddBoundaryCondition("0", "abs(y) < eps", params.Y)
	f.AddBoundaryCondition("0", "abs(z) < eps", params.Z)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func calcConsole4() {
	meshName, resName := "data/console4.mesh", "data/console4.res"
	f := fem.NewStaticFEM()
	err := f.SetMesh(meshName)
	if err != nil {
		fmt.Println("\nFEM error: ", err)
		return
	}
	f.SetNumThread(1)
	f.SetEps(1.0e-10)
	f.AddYoungModulus("203200", "")
	f.AddPoissonRatio("0.27", "")
	f.AddThickness("1", "")
	f.AddBoundaryCondition("0", "x == 0", params.X|params.Y)
	f.AddVolumeLoad("-1", "", params.Y)
	if err = f.Calculate(); err != nil {
		fmt.Println("\nFEM error:", err)
	} else if err = f.SaveResult(resName); err != nil {
		fmt.Println("\nFEM error:", err)
	}
}

func main() {
	//calcSquare()
	//calcConsole()
	//calcRail()
	//calcCube()
	//calcQuad()
	//calcShelTube()
	//calcShel()
	//calcTank()
	//calcConsole4()


	
	web.StartServer(9001)
}
