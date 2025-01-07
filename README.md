
# sd-ai
Proxy service to support compact prompts returning System Dynamics content.  

The intention is for this to be a free and public service hosted by isee systems and/or anyone else to do the engineering work of prompting LLM models for the purposes of generating CLDs and (eventually) quantitative SD models.  The service returns model information both as a JSON object of variables of relationships, and XMILE.  This service is what we at isee systems are using / will use in the future to build our LLM features around.  

The prompts in /engine/default/prompts/default use https://github.com/bear96/System-Dynamics-Bot as a point of departure.  

### To get started using this service....

1. npm i 
2. npm start 

We recommend VSCode using a launch.json for the Node type applications (you get a debugger, and hot-reloading)  

### How it works

The intent is to allow the community to build their own "engines" for doing SD model (or for the momment CLD only) generation using LLMs.  We provide a simple to implement interface that allows developers to create their own SD model generation engines or to extend and do research using the default OpenAI based engine which has been designed to be very flexible for modifiation without deep knowledge of coding.  The engine interface specifies everything Stella Architect v3.8 or greater needs to present a GUI to an end user and interact with any engine written by any memeber of the community.

In the default engine, the intent is to externalize all prompts, and LLM choices and make them availble as options for the end user. If the community desires we're open to supporting other APIs besides OpenAI.  This allows researchers in the field to do prompt engineering, and perfect the science around generating SD content from LLMs without having to worry about the engineering to make their work more generally available.  Stella Professional/Stella Architect or any other client which consumes this service will do the work of graph drawing, and user editing of returned models, allowing researchers within the field to focus purely on developing better ways to interact with LLMs. 

To make your own set of prompts, copy the /engines/default/prompts/default folder, make a new folder with the same 5 files (different content, but same names) and make the directory name something identifiable for the end user. When you do this be sure to preserve all JSON formatting information in the system prompt and the check (polarity) prompt as that is what the OpenAIWrapper class expects to get as a reply.  By making your own folder of prompts when you restart the server the user will be presented with an option to select your prompting scheme from a dropdown list.  

Likewise if you are a skilled JS developer you can write your own engine following the two example engines we have developed.  The first, fully featured engine is the default engine.  The second is a dummy predator prey engine which always returns the same content just as a simple demo.

The service can be run with an embedded OpenAI key (see note below) or an OpenAI key can be provided to each API call which interacts with OpenAI.  The Stella client has a place where a key can be provided, as well as a service address so that the Stella client can be pointed at any version of this service hosted by anyone else, including localhost for developers.

The service returns both a minimally viable XMILE representation of the model (no diagram information) which can be opened directly in Stella v3.7.3 or later and the view information will be machine generated by Stella, as well as a JSON object that contains an array of relationship information and an array of variables.  This JSON object is how state is maintained by the service.  

### API Documentation

Below are the four REST API calls this service support, written in the order they are typically made

1. GET /api/v1/initialize

This call can be skipped, but is useful for determining if your client is supported by the service.  

This call takes 2 optional query parameters

`clientProduct` - String - The product name of the client that is talking with the service (for support checks) (e.g. Stella Architect).  
`clientVersion` - String - The version number of the client that is talking with the service (for support checks) (e.g. 3.8.0).  

Returns `{success: <bool>, message: <string> }`

2. GET /api/v1/engines

This call can be skipped, but is useful if your client wants to know what engines are available.

This call takes no query parameters

Returns `{success: <bool>, engines:['default', 'predprey', .... any other engines in the /engines folder] }`

3. GET /api/v1/engines/:engine/parameters

This call can be skipped, but is useful if your client wants to know what parameters a particular engine supports

This call takes no query parameters

Returns 
```
{ success: <bool>, 
    parameters:[{
        name: <string, unique name for the parmater that is passed to generate call>,
        type: <string, currently this service only supports 'string' for this attribute>,
        required: <boolean, whether or not this parameter must be passed to the generate call>,
        uiElement: <string, type of UI element the client should use so that the user can enter this value.  Valid values are textarea|lineedit|password|combobox|hidden>,
        label: <string, name to put next to the UI element in the client>,
        description: <string, description of what this parameter does, used as a tooltip or placeholder text in the client>,
        defaultValue: <string, default value for this parameter if there is one, otherwise skipped>,
        options: <array, of objects with two attributes 'label' and 'value' only used if 'uiElement' is combobox>,
        saveForUser: <string, whether or not this field should be saved for the user by the client, valid values are local|global leave unspecified if not saved>,
    }] 
}
```

4. POST /api/v1/:engine/generate

This call is the work-horse of the service, doing the job of diagram generation.

This call takes at least 3 post parameters, all other parameters are found via a call to /api/v1/:engine/parameters

`prompt` - The prompt typed in by the end user
`format` - The return type for the information. Either sd-json or xmile, default is sd-json
`currentModel` - The sd-json representation of the current model as JSON.

Returns `{success: <bool>, message: <string>, format: <string>, model: {variables: [], relationships: []}}`  

All relationships in the entire diagram will be returned irregardless of whether they are new or not.  The client is expected to do any diff/update operations if desired.  The "normal" usecase is that each call to generate returns a whole "new" model.

sd-json format is:
```
{
    variables: [{
        name: <string>,
        type: <string - stock|flow|variable>
    }], 
    relationships: [{
        "reasoning": <string, explanation for why this relationship is here> 
        "causalRelationship": <string, "source --> destination">,  
        "relevantRext": <string, portion of the prompt or research for why this relationship is here>, 
        "polarity": <string ?|+|- >, 
        "polarityReasoning": <string explanation for why this polarity was chosen> 
    }]
}
```  


### Important Note 
You must have a .env file at the top level which can have the following keys  
 * OPENAI_API_KEY which is your open AI access token, if provided then clients do not need to provide one to either the intialize or generate calls  