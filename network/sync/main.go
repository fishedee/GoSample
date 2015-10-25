package main;

import (
	"fmt"
	"time"
	"os"
	"bufio"
	"io"
	"bytes"
	"strconv"
	"errors"
	"strings"
	"net/http"
	"net/url"
	"net/http/cookiejar"
	"io/ioutil"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
)

type Config struct{
	TargetPhone string
	TargetClassify string
	TargetName string
	SourceRecipeId int
}

type RecipeMaterial struct{
	Name string `json:"name"`
	Weight string `json:"weight"`
}

type RecipeStep struct{
	Image string `json:"image"`
	Text string `json:"text"`
}

type Recipe struct{
	ContentId string `json:"contentId"`
	Title string `json:"title"`
	Image string `json:"image"`
	Summary string `json:"summary"`
	Material []RecipeMaterial `json:"material"`
	Step []RecipeStep `json:"step"`
	Tip string `json:"tip"`
}

type JsonRecipe2 struct{
	Recipe *Recipe `json:"recipe"`
}

type JsonRecipe struct{
	Code int `json:"cod"`
	Msg string `json:"msg"`
	Data JsonRecipe2 `json:"data"`
}

func Atoi(s string)(int){
	result,error := strconv.Atoi(s);
	if error != nil{
		return 0;
	}
	return result;
}

func Itoa(i int)(string){
	result := strconv.Itoa(i);
	return result;
}

func getConfigFromTxt( fileName string )([]*Config,error){
	file, error := os.Open(fileName)
    if error != nil {
    	return nil,error
    }
    defer file.Close()

    readFd := bufio.NewReader(file)
    var result []*Config;
    for {
        line, error := readFd.ReadString('\n') //以'\n'为结束符读入一行
        
        if error == io.EOF{
        	break;
        }
        if error != nil{
        	return nil,error;
        }

        lineSplit := strings.Split(line," ");
        if len(lineSplit) != 4 {
        	return nil,errors.New("行不是四列"+line);
        }

        result = append(
        	result,
        	&Config{
        		TargetPhone:lineSplit[0],
        		TargetClassify:lineSplit[1],
        		TargetName:lineSplit[3],
        		SourceRecipeId:Atoi(lineSplit[2]),
        	},
        );
	}   

	return result,nil;
}

func sync(config *Config)(error){
	data,error := getRecipeFromHonbeibang(config.SourceRecipeId);
	if error != nil{
		return error;
	}

	error = uploadRecipeToThere(data,config.TargetPhone,config.TargetClassify,config.TargetName);
	if error != nil{
		return error;
	}

	return nil;
}

func getRecipeFromHonbeibang(contentId int)(*Recipe,error){
	resp, error := http.Get("http://www.hongbeibang.com/recipe/get?contentId="+Itoa(contentId))
	if error != nil {
		return nil,error;
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil,errors.New("返回码不是200 "+Itoa(resp.StatusCode));
	}

	body,error := ioutil.ReadAll(resp.Body)
	if error != nil{
		return nil,error;
	}
	
	var result *JsonRecipe;
	error = json.Unmarshal(body,&result);
	if error != nil{
		return nil,error;
	}

	return result.Data.Recipe,nil;
}

func getThereXSRFToken(client *http.Client,url string,query string)(string,error){
	reqest,error := http.NewRequest("GET",url,nil);
	if error != nil{
		return "",error;
	}

	response,error := client.Do(reqest)
	if error != nil{
		return "",error;
	}
	if response.StatusCode != 200 {
		return "",errors.New("返回码不是200 "+Itoa(response.StatusCode)+","+url);
	}

	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
	    return "",error;
	}

	return  doc.Find(query).AttrOr("value",""),nil;
}

func postFormThere(client *http.Client,postUrl string,args map[string]string)(string,error){
	form := url.Values{}
	for key,value := range args{
		form.Add(key,value)
	}

	reqest,error := http.NewRequest("POST",postUrl, bytes.NewBufferString(form.Encode()));
	if error != nil{
		return "",error;
	}

	reqest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    reqest.Header.Add("Content-Length", Itoa(len(form.Encode())))

	response,error := client.Do(reqest)
	if error != nil{
		return "",error;
	}
	if response.StatusCode != 200 {
		return "",errors.New("返回码不是200 "+Itoa(response.StatusCode)+","+postUrl);
	}

	body,error := ioutil.ReadAll(response.Body)
	if error != nil{
		return "",error;
	}

	return string(body),nil;
}
func loginThere(client *http.Client)(error){
	token,error := getThereXSRFToken(client,"http://lamsoon.solomochina.com/admin/login","form input[name=_token]");
	if error != nil{
		return error;
	}
	fmt.Println(token);
	args := map[string]string{
		"username":"admin",
		"password":"1",
		"_token":token,
	}
	result,error := postFormThere(client,"http://lamsoon.solomochina.com/admin/login",args);
	if error != nil{
		return error;
	}

	fmt.Println(result);
	return nil;
}

func uploadRecipeToThere(recipe *Recipe,phone string,classify string,name string)(error){
    jar, error := cookiejar.New(nil)
    if error != nil {
    	return error;
    }
	client := &http.Client{Jar:jar}

	loginThere(client);
	return nil;
}

func main(){
	config,error := getConfigFromTxt("sync.txt");
	if error != nil{
		fmt.Println("读取配置文件失败");
		return;
	}
	fmt.Println("配置文件数据量：",len(config));

	for _,singleConfig := range config{
		fmt.Println("开始转换食谱ID ",singleConfig.SourceRecipeId," ...");
		error = sync( singleConfig );
		if error != nil{
			fmt.Println("转换食谱ID ",singleConfig.SourceRecipeId," 失败：",error.Error());
		}else{
			fmt.Println("转换食谱ID ",singleConfig.SourceRecipeId," 成功");
		}
		<-time.After(0 * time.Minute)
	}
}