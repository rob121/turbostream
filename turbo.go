package turbostream

import("fmt")


/*
ppend, prepend, (insert) before, (insert) after, replace, update, and remove
*/

func Message(action string, target string,content string) ([]byte){


	if(!actionValid((action))){

		out := fmt.Sprintf(`<turbo-stream action='replace' target='%s'>
	<template>INVALID ACTION SET</template>
	</turbo-stream>`,action,target,content)

		return []byte(out)

	}

	out := fmt.Sprintf(`<turbo-stream action='%s' target='%s'>
	<template>%s</template>
	</turbo-stream>`,action,target,content)
	return []byte(out)
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