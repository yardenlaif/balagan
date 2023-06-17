package main

import (
	"fmt"
	"go/types"
	"reflect"

	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/analysis"
)

var interfaces []*types.Package

type interfaceFactType struct {
	interfaces map[*types.Interface]struct{}
}

func (i *interfaceFactType) AFact() {}

var InterfaceFinder = analysis.Analyzer{
	Name:       "InterfaceFinder",
	Doc:        "Find all interfaces",
	FactTypes:  []analysis.Fact{&interfaceFactType{}},
	ResultType: reflect.TypeOf(map[*types.Interface]struct{}{}),
	Run: func(p *analysis.Pass) (interface{}, error) {
		fmt.Printf("p: %v\n", p.Pkg.Path())
		interfaces := make(map[*types.Interface]struct{})
		for _, i := range findInterfaces(p.TypesInfo) {
			interfaces[i] = struct{}{}
		}
		for _, i := range p.Pkg.Imports() {
			var fact interfaceFactType
			p.ImportPackageFact(i, &fact)
			maps.Copy(interfaces, fact.interfaces)
		}
		p.ExportPackageFact(&interfaceFactType{interfaces: interfaces})
		return interfaces, nil
	},
}

type implementationMap map[*types.Func][]*types.Func

func (i *implementationMap) AFact() {}

var InterfaceImplementationFinder = analysis.Analyzer{
	Name:       "InterfaceImplementationFinder",
	Doc:        "Find all interface implementations",
	FactTypes:  []analysis.Fact{&implementationMap{}},
	ResultType: reflect.TypeOf(&implementationMap{}),
	Run: func(p *analysis.Pass) (interface{}, error) {
		implementations := make(map[*types.Func][]*types.Func)
		for _, i := range p.Pkg.Imports() {
			var fact implementationMap
			p.ImportPackageFact(i, &fact)
			maps.Copy(implementations, fact)
		}

		interfaces := p.ResultOf[&InterfaceFinder].(map[*types.Interface]struct{})
		for _, def := range p.TypesInfo.Defs {
			if fun, ok := def.(*types.Func); ok {
				recv := fun.Type().(*types.Signature).Recv()
				if recv == nil {
					continue
				}

				for i := range interfaces {
					if !types.Implements(recv.Type(), i) {
						continue
					}

					obj, _, _ := types.LookupFieldOrMethod(i, false, recv.Pkg(), fun.Name())
					if method, ok := obj.(*types.Func); ok {
						implementations[method] = append(implementations[method], fun)
					}
				}
			}
		}

		p.ExportPackageFact((*implementationMap)(&implementations))
		return (*implementationMap)(&implementations), nil
	},
	Requires: []*analysis.Analyzer{&InterfaceFinder},
}

func findInterfaces(info *types.Info) []*types.Interface {
	interfaces := make(map[*types.Interface]struct{})
	for _, typeAndValue := range info.Types {
		t := typeAndValue.Type
		if named, ok := t.(*types.Named); ok {
			t = named.Underlying()
		}
		if n, ok := t.(*types.Interface); ok {
			interfaces[n] = struct{}{}
		}
	}
	return maps.Keys(interfaces)
}

func findImplementations(info *types.Info) map[*types.Func][]*types.Func {
	implementations := make(map[*types.Func][]*types.Func)
	interfaces := findInterfaces(info)
	for _, def := range info.Defs {
		if fun, ok := def.(*types.Func); ok {
			recv := fun.Type().(*types.Signature).Recv()
			if recv == nil {
				continue
			}

			for _, i := range interfaces {
				if !types.Implements(recv.Type(), i) {
					continue
				}

				obj, _, _ := types.LookupFieldOrMethod(i, false, recv.Pkg(), fun.Name())
				if method, ok := obj.(*types.Func); ok {
					implementations[method] = append(implementations[method], fun)
				}
			}
		}
	}
	return implementations
}
