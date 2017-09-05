package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func findFiles(root string, patterns []string, excludes []string) (files []string) {
	fmt.Println("root " + root)
	files = make([]string, 0)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		ok := false
		for _, e := range excludes {
			if strings.Contains(path, e) {
				fmt.Println("except " + path)
				return nil
			}
		}
		for _, p := range patterns {
			ok, _ = filepath.Match(p, filepath.Base(path))
			if ok {
				break
			}
		}
		if ok {
			fmt.Println("add file " + path)
			files = append(files, path)
		}
		return nil
	})
	return files
}

type ProjectFile struct {
	Name   string
	Filter string
}

type Project struct {
	Name           string
	Filters        map[string]bool
	Files          []*ProjectFile
	PropertySheets []string
	UUID			string
}

func main() {

	root := flag.String("root", ".", "root")
	name := flag.String("name", "agame", "name of project")
	exclude := flag.String("exclude", "", "exclude keyword")
	prjFile := flag.String("prjFile", "agame.vcxproj", "project file")
	filterFile := flag.String("filterFile", "agame.vcxproj.filters", "project filter file")
	mapRoot := flag.String("mapRoot", "", "map root")
	propertySheets := flag.String("p", "", "property sheets, seperated by comma")
	filePatterns := flag.String("filePatterns", "*.h,*.cpp,*.c", "files, seperated by comma")
	uuidStr := flag.String("uuid", "6757D22F-2E69-4E79-BDC3-A7F8B8FC7471", "project uuid")
	flag.Parse()

	rootAbs, _ := filepath.Abs(*root)

	filePatternArr := strings.Split(*filePatterns, ",")
	excludes := make([]string, 0)
	if *exclude != "" {
		excludes = strings.Split(*exclude, ",")
	}

	files := findFiles(*root, filePatternArr, excludes)

	prj := &Project{
		Filters: make(map[string]bool),
		Files:   make([]*ProjectFile, 0),
		Name:    *name,
		UUID:	 *uuidStr,
	}

	for _, file := range files {
		name := file[len(rootAbs)+1:]
		dir := filepath.Dir(name)

		if *mapRoot != "" {
			file = *mapRoot + "/" + name
			file = strings.Replace(file, "/", "\\", -1)
			dir = strings.Replace(dir, "/", "\\", -1)
		}
		prjFile := &ProjectFile{
			Name:   file,
			Filter: dir,
		}
		prj.Files = append(prj.Files, prjFile)
		for {
			prj.Filters[dir] = true
			pos := strings.LastIndex(dir, "\\")
			if pos == -1 {
				break
			}
			dir = dir[:pos]
		}
	}

	prj.PropertySheets = strings.Split(*propertySheets, ",")
	fmt.Println(prj.PropertySheets)

	buf := &bytes.Buffer{}
	prjTpl := template.Must(template.New("vsxproj").Parse(vsxproj))
	prjTpl.Execute(buf, prj)
	ioutil.WriteFile(*prjFile, buf.Bytes(), 0644)

	buf = &bytes.Buffer{}
	filterTpl := template.Must(template.New("filter").Parse(filter))
	filterTpl.Execute(buf, prj)
	ioutil.WriteFile(*filterFile, buf.Bytes(), 0644)

	fmt.Println(prjTpl, filterTpl)
}

var vsxproj = `<?xml version="1.0" encoding="utf-8"?>
<Project DefaultTargets="Build" ToolsVersion="15.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup Label="ProjectConfigurations">
    <ProjectConfiguration Include="Debug|Win32">
      <Configuration>Debug</Configuration>
      <Platform>Win32</Platform>
    </ProjectConfiguration>
  </ItemGroup>
  <PropertyGroup Label="Globals">
    <ProjectGuid>{{"{"}}{{.UUID}}{{"}"}}</ProjectGuid>
    <RootNamespace>{{.Name}}</RootNamespace>
    <WindowsTargetPlatformVersion>8.1</WindowsTargetPlatformVersion>
  </PropertyGroup>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.Default.props" />
  <PropertyGroup Condition="'$(Configuration)|$(Platform)'=='Debug|Win32'" Label="Configuration">
    <ConfigurationType>Application</ConfigurationType>
    <UseDebugLibraries>true</UseDebugLibraries>
    <PlatformToolset>v150</PlatformToolset>
    <CharacterSet>MultiByte</CharacterSet>
  </PropertyGroup>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.props" />
  <ImportGroup Label="ExtensionSettings">
  </ImportGroup>
  <ImportGroup Label="Shared">
  </ImportGroup>
{{ range .PropertySheets }}
  <ImportGroup Label="PropertySheets">
    <Import Project="{{ . }}" />
  </ImportGroup>
{{ end }}
  <ImportGroup Label="PropertySheets" Condition="'$(Configuration)|$(Platform)'=='Debug|Win32'">
    <Import Project="$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props" Condition="exists('$(UserRootDir)\Microsoft.Cpp.$(Platform).user.props')" Label="LocalAppDataPlatform" />
  </ImportGroup>
  <PropertyGroup Label="UserMacros" />
  <PropertyGroup />
  <ItemDefinitionGroup Condition="'$(Configuration)|$(Platform)'=='Debug|Win32'">
    <ClCompile>
      <WarningLevel>Level3</WarningLevel>
      <Optimization>Disabled</Optimization>
      <SDLCheck>true</SDLCheck>
    </ClCompile>
  </ItemDefinitionGroup>
  <ItemGroup>
{{ range .Files }}
    <ClCompile Include="{{.Name}}">
    </ClCompile>
{{ end }}
  </ItemGroup>
  <Import Project="$(VCTargetsPath)\Microsoft.Cpp.targets" />
  <ImportGroup Label="ExtensionTargets">
  </ImportGroup>
</Project>
`

var filter = `<?xml version="1.0" encoding="utf-8"?>
<Project ToolsVersion="4.0" xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <ItemGroup>
{{ range $key, $_ := .Filters }}
    <Filter Include="{{$key}}">
    </Filter>
{{ end }}
  </ItemGroup>
  <ItemGroup>
{{ range .Files }}
    <ClCompile Include="{{ .Name }}">
      <Filter>{{.Filter}}</Filter>
    </ClCompile>
{{ end }}
  </ItemGroup>
</Project>
`
