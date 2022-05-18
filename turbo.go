package turbostream

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
)


var logger *log.Logger
var defaultChannel string

func init(){

	 defaultChannel  = "allclients"
	 logger = log.New(ioutil.Discard, "[turbostream]", log.Lshortfile)

}

func Logger(l *log.Logger){
       logger = l
}

/*
append, prepend, (insert) before, (insert) after, replace, update, and remove
*/

func Message(action string, target string,content string) ([]byte){


	if(!actionValid((action))){

		out := fmt.Sprintf(`<turbo-stream action='replace' target='%s'>
	<template>INVALID ACTION SET</template>
	</turbo-stream>`,target)

		return []byte(out)

	}

	out := fmt.Sprintf(`<turbo-stream action='%s' target='%s'>
	<template>%s</template>
	</turbo-stream>`,action,target,content)
	return []byte(out)
}

func MessageTmpl(action string, target string,tmpl *template.Template,data map[string]interface{}) ([]byte,error){


	if(!actionValid((action))){

		out := fmt.Sprintf(`<turbo-stream action='replace' target='%s'>
	<template>INVALID ACTION SET</template>
	</turbo-stream>`,target)

		return []byte(out),errors.New("INVALID ACTION")

	}

	var tpl bytes.Buffer

	if err := tmpl.Execute(&tpl, data); err != nil {

		out := fmt.Sprintf(`<turbo-stream action='replace' target='%s'>
	<template>Template Parse Error %s</template>
	</turbo-stream>`,target,err.Error())

		return []byte(out),err

	}

	out := fmt.Sprintf(`<turbo-stream action='%s' target='%s'>
	<template>%s</template>
	</turbo-stream>`,action,target,tpl.String())

	return []byte(out),nil
}

func actionValid(action string) (bool){

	valid := []string{"append", "prepend","before","after","replace","update","remove"}


    for _,v := range valid {


    	if(v == action){

     	return true
	    }

	}

	return false


}