//*****************************************************************
// Program: readYAML
// Purpose: Read data from yaml files and create endpoints in consul
// Author: SidhuG
//*******************************************************************

package main

import (
    "fmt"
    "os"
    //"strings"
    //"text/template"
    "reflect"
    "gopkg.in/yaml.v2"
)

const data = `
Colors:
  - red: red
  - pink:
     mix:
      - white
      - red
  - blue: blue
  - white: white
`
//type Metadata struct {
    // Keys are the keys of the structure which were successfully decoded
  //  Keys []string

    // Unused is a slice of keys that were found in the raw value but
    // weren't decoded since there was no matching field in the result interface
//    Unused []string
//}

type I interface{}
type A struct {
    Greeting string
    Message  string
    Pi       float64
}

type B struct {
    Struct    A
    Ptr       *A
    Answer    int
    Map       map[string]string
    StructMap map[string]interface{}
    Slice     []string
}

func main() {
    
    m := make(map[interface{}]interface{})
    //dataRead := make(map[interface{}]interface{})
    
    err := yaml.Unmarshal([]byte(data), &m)
    checkError(err)
    fmt.Printf("--- m:\n%v\n\n", m)

    //var valueType reflect.kind

    fmt.Println("-----Printing toplevel map-----")
    for k, v := range m { 
        fmt.Printf("key[%s] value[%s]\n", k, v)
        valueType := reflect.TypeOf(v).Kind()
        fmt.Printf("ValueType is %s", valueType)
        translate(v)
        fmt.Println()

        //if valueType == Map {
            //vmap := k[v].([]interface{})
        //    fmt.Printf("Value is Map")
        //}

    }

    //the_list := y["list"].([]interface{})
    
    // Assume we have a struct that contains an interface of an unknown type
    //fmt.Println("Translating a struct containing a pointer to a struct wrapped in an interface:")
    //type D struct {
    //    Payload *I
    //}
    //original := D{
    //    Payload: &m,
    //}
    //translated := translate(original)
    //fmt.Println("original:  ", original, "->", (*original.Payload), "->", (*original.Payload).(B).Ptr)
    //fmt.Println("translated:", translated, "->", (*translated.(D).Payload), "->", (*(translated.(D).Payload)).(B).Ptr)
    //fmt.Println()


}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}

func translate(obj interface{}) interface{} {
    // Wrap the original in a reflect.Value
    original := reflect.ValueOf(obj)

    copy := reflect.New(original.Type()).Elem()
    translateRecursive(copy, original)

    // Remove the reflection wrapper
    return copy.Interface()
}

func translateRecursive(copy, original reflect.Value) {

    switch original.Kind() {
    // The first cases handle nested structures and translate them recursively

    // If it is a pointer we need to unwrap and call once again
    case reflect.Ptr:
        // To get the actual value of the original we have to call Elem()
        // At the same time this unwraps the pointer so we don't end up in
        // an infinite recursion
        originalValue := original.Elem()
        // Check if the pointer is nil
        if !originalValue.IsValid() {
            return
        }
        // Allocate a new object and set the pointer to it
        copy.Set(reflect.New(originalValue.Type()))
        // Unwrap the newly created pointer
        translateRecursive(copy.Elem(), originalValue)

    // If it is an interface (which is very similar to a pointer), do basically the
    // same as for the pointer. Though a pointer is not the same as an interface so
    // note that we have to call Elem() after creating a new object because otherwise
    // we would end up with an actual pointer
    case reflect.Interface:
        // Get rid of the wrapping interface
        originalValue := original.Elem()
        // Create a new object. Now new gives us a pointer, but we want the value it
        // points to, so we have to call Elem() to unwrap it
        copyValue := reflect.New(originalValue.Type()).Elem()
        translateRecursive(copyValue, originalValue)
        copy.Set(copyValue)

    // If it is a struct we translate each field
    case reflect.Struct:
        for i := 0; i < original.NumField(); i += 1 {
            translateRecursive(copy.Field(i), original.Field(i))
        }

    // If it is a slice we create a new slice and translate each element
    case reflect.Slice:
        copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
        for i := 0; i < original.Len(); i += 1 {
            translateRecursive(copy.Index(i), original.Index(i))
        }

    // If it is a map we create a new map and translate each value
    case reflect.Map:
        copy.Set(reflect.MakeMap(original.Type()))
        for _, key := range original.MapKeys() {
            fmt.Println()
            fmt.Printf(" Key: %s  -> ",key )
            originalValue := original.MapIndex(key)
            // New gives us a pointer, but again we want the value
            copyValue := reflect.New(originalValue.Type()).Elem()
            translateRecursive(copyValue, originalValue)
            copy.SetMapIndex(key, copyValue)
        }

    // Otherwise we cannot traverse anywhere so this finishes the the recursion

    // If it is a string translate it (yay finally we're doing what we came for)
    case reflect.String:
        translatedString := original.Interface().(string)
        copy.SetString(translatedString)
        fmt.Printf(" Value:   %s ", translatedString)

    // And everything else will simply be taken from the original
    default:
        copy.Set(original)
    }
}
