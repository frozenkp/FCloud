package main

import(
  "gopkg.in/gin-gonic/gin.v1"
  "net/http"
  "os"
  "io"
  "encoding/csv"
  "hash/fnv"
  "strconv"
  "os/exec"
  "strings"
  "log"
)

type File struct{
  Owner string
  Name  string
  Size  string
}

var users map[string]string

func initial(){
  users=make(map[string]string)
  //get user lists
  f,_:=os.OpenFile("./users/userlist",os.O_RDONLY,0777)
  defer f.Close()
  r:=csv.NewReader(f)
  for true {
    info,err:=r.Read()
    if err==io.EOF {
      break
    }else{
      users[info[0]]=info[1]
    }
  }
}

func hash(s string)string{
  h := fnv.New32a()
  h.Write([]byte(s))
  return strconv.Itoa(int(h.Sum32()))
}


func login(c *gin.Context){
  user:=c.PostForm("username")
  pswd:=hash(c.PostForm("password"))
  if users[user]==pswd {
    /*cookie := http.Cookie{Name: "user", Value: pswd, Path: "/", MaxAge: 1800}
    http.SetCookie(c.Writer,&cookie)*/
    c.SetCookie("user",pswd,1800,"/","",false,true)
    c.Redirect(http.StatusMovedPermanently,"/user/"+user+"/home")
  }else{
    c.Writer.Write([]byte("<script>alert(\"帳號或是密碼錯誤.\")</script>"))
    c.Writer.Write([]byte("<script>document.location.href=\"/\";</script>"))
  }
}

func home(c *gin.Context){
  user:=c.Param("name")
  pswd,err:=c.Cookie("user")
  //check user match
  if users[user]==pswd && err==nil {
    //renew cookie
    c.SetCookie("user",pswd,1800,"/","",false,true)
    //get file list
    list,_:=exec.Command("sh","script/list.sh","./users/"+user+"/files").Output()
    files:=strings.Split(string(list),"\n")
    f:=make([]File,0)
    for i:=0;i<len(files)-1;i++ {
      info:=strings.Split(files[i]," ")
      f=append(f,File{user,info[0],info[1]})
    }
    //write html
    c.HTML(http.StatusOK,"drive.tmpl",gin.H{"User":user, "Files":f})
  }else{
    c.Redirect(http.StatusMovedPermanently,"/")
  }
}

func del(c *gin.Context){
  user:=c.Param("name")
  pswd,err:=c.Cookie("user")
  file:=c.Param("filename")
  //check user match
  if users[user]==pswd && err==nil {
    //renew cookie
    c.SetCookie("user",pswd,1800,"/","",false,true)
    //delete file
    err:=os.Remove("./users/"+user+"/files/"+file)
    if err!=nil {
      log.Print(err)
    }
    c.Redirect(http.StatusMovedPermanently,"/user/"+user+"/home")
  }else{
    c.Redirect(http.StatusMovedPermanently,"/")
  }
}

func upload(c *gin.Context){
  user:=c.Param("name")
  pswd,err:=c.Cookie("user")
  //check user match
  if users[user]==pswd && err==nil {
    //renew cookie
    c.SetCookie("user",pswd,1800,"/","",false,true)
    //upload file
    file, header, err:=c.Request.FormFile("file")
    filename:=header.Filename
    out, err := os.Create("./users/"+user+"/files/"+filename)
    if err != nil {
      log.Print(err)
    }
    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
      log.Print(err)
    }
    c.Redirect(http.StatusMovedPermanently,"/user/"+user+"/home")
  }else{
    c.Redirect(http.StatusMovedPermanently,"/")
  }
}

func download(c *gin.Context){
  user:=c.Param("name")
  pswd,err:=c.Cookie("user")
  filename:=c.Param("filename")
  //check user match
  if users[user]==pswd && err==nil {
    //renew cookie
    c.SetCookie("user",pswd,1800,"/","",false,true)
    //download link
    in,_:=os.Open("./users/"+user+"/files/"+filename)
    defer in.Close()
    //c.Header("Content-Type","application/octet-stream")
    c.Header("Content-Type",c.Request.Header.Get("Content-Type"))
    c.Header("Content-Disposition", "attachment; filename="+filename)
    c.Header("Content-Length",c.Request.Header.Get("Content-Length"))
    c.Header("Content-Transfer-Encoding","binary")
    //write file to w
    io.Copy(c.Writer,in)
  }else{
    c.Redirect(http.StatusMovedPermanently,"/")
  }
}

func main(){
  initial()

  router := gin.Default()

  router.LoadHTMLGlob("./tmpl/*")

  router.Static("/js","./js")
  router.Static("/css","./css")
  router.Static("/images","./images")
  router.StaticFile("/","./index.html")

  router.POST("/login", login)
  router.GET("/user/:name/home", home)
  router.GET("/user/:name/file/delete/:filename", del)
  router.GET("/user/:name/file/download/:filename", download)
  router.POST("/user/:name/file/upload", upload)

  router.Run(":8080")
}
