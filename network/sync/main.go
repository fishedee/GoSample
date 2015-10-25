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
	"mime/multipart"
	"path/filepath"
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

func getThereXSRFToken(data string,query string)(string,error){
	doc, error := goquery.NewDocumentFromReader(bytes.NewBufferString(data))
	if error != nil {
	    return "",error;
	}

	return  doc.Find(query).AttrOr("value",""),nil;
}

func handleRequest(client *http.Client,reqest *http.Request,contentType string)(string,error){
	reqest.Header.Add("Content-Type", contentType)
    //reqest.Header.Add("Content-Length",Itoa(contentLength) )
    reqest.Header.Add("User-Agent","Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36")

	response,error := client.Do(reqest)
	if error != nil{
		return "",error;
	}
	if response.StatusCode != 200 {
		return "",errors.New("返回码不是200 "+Itoa(response.StatusCode));
	}

	body,error := ioutil.ReadAll(response.Body)
	if error != nil{
		return "",error;
	}

	return string(body),nil;
}
func apiThere(client *http.Client,method string,postUrl string,args map[string]string)(string,error){
	form := url.Values{}
	if args != nil{
		for key,value := range args{
			form.Add(key,value)
		}
	}

	reqest,error := http.NewRequest(method,postUrl, bytes.NewBufferString(form.Encode()));
	if error != nil{
		return "",error;
	}

	return handleRequest(client,reqest,"application/x-www-form-urlencoded");
}

func apiFileThere(client *http.Client,method string,postUrl string,args map[string]string,fileParam string,fileAddressArray []string)(string,error){
	bodyBuf := &bytes.Buffer{}
    bodyWriter := multipart.NewWriter(bodyBuf)

    for _,fileAddress := range fileAddressArray{
    	if fileAddress == ""{
    		bodyWriter.CreateFormFile(fileParam,"")
		}else{
			fileWriter, error := bodyWriter.CreateFormFile(fileParam, filepath.Base(fileAddress))
		    if error != nil{
		    	return "",error
		    }
		    file, err := os.Open(fileAddress)
			if err != nil {
			  	return "", err
			}
			defer file.Close()
		    _, error = io.Copy(fileWriter, file)
		    if error != nil{
		    	return "",error
		    }
		}
    }
    
    for key,value := range args{
    	error := bodyWriter.WriteField(key,value)
    	if error != nil{
    		return "",error
    	}
    }

    reqest,error := http.NewRequest(method,postUrl,bodyBuf);
	if error != nil{
		return "",error;
	}

	return handleRequest(client,reqest,bodyWriter.FormDataContentType());
}

func loginThere(client *http.Client)(error){
	//获取登录页面的token
	result,error := apiThere(client,"GET","http://lamsoon.solomochina.com/admin/login",nil)
	if error != nil{
		return error
	}

	token,error := getThereXSRFToken(result,"form input[name=_token]");
	if error != nil{
		return error;
	}

	//登录
	args := map[string]string{
		"username":"admin",
		"password":"123",
		"_token":token,
	}
	_,error = apiThere(client,"POST","http://lamsoon.solomochina.com/admin/login",args);
	if error != nil{
		return error;
	}

	return nil;
}

func loginVirtual(client* http.Client,phone string)(error){
	//获取水军页面的token
	result,error := apiThere(client,"GET","http://lamsoon.solomochina.com/admin/publish/login",nil)
	if error != nil{
		return error
	}

	token,error := getThereXSRFToken(result,"form input[name=_token]");
	if error != nil{
		return error;
	}
	
	//登录水军
	args := map[string]string{
		"mobile":phone,
		"_token":token,
	}
	_,error = apiThere(client,"POST","http://lamsoon.solomochina.com/admin/publish/login",args);
	if error != nil{
		return error;
	}

	return nil;
}

type Topic struct{
	CategoryId int `json:"category_id"`
	Id int `json:"id"`
	Title string `json:"title"`
	UserName string `json:"user_name"`
}

type JsonTopicList struct{
	CurrentPage int `json:"current_page"`
	Data []Topic `json:"data"`
}

type JsonTopic struct{
	TopicList JsonTopicList `json:"topic_list"`
}


func getTopicId(client* http.Client,title string)(int,error){
	for i := 1 ; i <= 5 ; i++{
		result,error := apiThere(client,"GET","http://lamsoon.solomochina.com/api/topic?page=1&top_category_id=6",nil)
		if error != nil{
			return 0,error
		}

		var jsonTopic *JsonTopic;
		error = json.Unmarshal([]byte(result),&jsonTopic)
		if error != nil{
			return 0,error
		}

		for _,value := range jsonTopic.TopicList.Data{
			if value.Title == title{
				return value.Id,nil;
			}
		}
	}
	return 0,errors.New("找不到食谱对应的ID"+title)
}

func postTopic(client* http.Client,categoryId int,title string,content string,image string)(error){
	//获取发帖页面的token
	result,error := apiThere(client,"GET","http://lamsoon.solomochina.com/admin/publish/add_topic",nil)
	if error != nil{
		return error
	}

	token,error := getThereXSRFToken(result,"form input[name=_token]");
	if error != nil{
		return error;
	}

	//发帖
	files := []string{}
	files = append(files,image)
	for i := 1 ; i != 9 ; i++{
		files = append(files,"")
	}

	args := map[string]string{
		"category_id":Itoa(categoryId),
		"title":title,
		"content":content,
		"_token":token,
	}
	_,error = apiFileThere(client,"POST","http://lamsoon.solomochina.com/admin/publish/add_topic",args,"photos[]",files);
	if error != nil{
		return error;
	}

	return nil;
}

func postComment(client* http.Client,topicId int,content string,image string)(error){
	//获取评论页面的token
	result,error := apiThere(client,"GET","http://lamsoon.solomochina.com/admin/publish/add_comment",nil)
	if error != nil{
		return error
	}

	token,error := getThereXSRFToken(result,"form input[name=_token]");
	if error != nil{
		return error;
	}

	//评论
	files := []string{}
	files = append(files,image)
	for i := 1 ; i != 9 ; i++{
		files = append(files,"")
	}

	args := map[string]string{
		"topic_id":Itoa(topicId),
		"content":content,
		"_token":token,
	}
	_,error = apiFileThere(client,"POST","http://lamsoon.solomochina.com/admin/publish/add_comment",args,"photos[]",files);
	if error != nil{
		return error;
	}

	return nil;
}

func uploadRecipeToThere(recipe *Recipe,phone string,classify string,name string)(error){
    jar, error := cookiejar.New(nil)
    if error != nil {
    	return error;
    }
	client := &http.Client{Jar:jar}

	error = loginThere(client);
	if error != nil{
		return error
	}

	error = loginVirtual(client,"13988888888");
	if error != nil{
		return error
	}

	error = postTopic(client,9,"标题1","内容1","test.jpg")
	if error != nil{
		return error
	}

	id,error := getTopicId(client,"标题1")
	if error != nil{
		return error
	}

	error = postComment(client,id,"回复1\n测试2","test.jpg");
	if error != nil{
		return error
	}

	error = postComment(client,id,"回复3\n测试4","test.jpg");
	if error != nil{
		return error
	}
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