package main

import(
	"os"
	"io"
	"log"
	"time"
	"path"
	"strings"
	"strconv"
	"net/http"
	"math/rand"
	"io/ioutil"
	"archive/zip"
	"path/filepath"
)
type Question struct {
	number int
	title string
	choice string
	answer string
}

type QuestionData struct {
	origin []int
	answer []string
	json string
	mark string
}

type StudentInfo struct {
	ID string
	name string
	choice QuestionData
	judgment QuestionData
	completion QuestionData
	answer QuestionData
	practice QuestionData
	cheat1 int
	cheat2 int
	cheat3 int
}

type ExamInfo struct {
	name string
	startTime int64
	endTime int64
	state string
	choiceCount int
	judgmentCount int
	completionCount int
	answerCount int
	practiceCount int
}

var(
	AllChoice []Question //选择题
	AllJudgment []Question //判断题
	AllCompletion []Question //填空题
	AllAnswer []Question //简答题
	AllPractice []Question //操作题
	AllStudent []StudentInfo
	examInfo ExamInfo
	examFlag bool
)

func main(){
	Init()
	examFlag = false
	http.HandleFunc("/", index)
	http.HandleFunc("/GetFile",GetFile)
	http.HandleFunc("/connectTest",connectTest)
	http.HandleFunc("/GetInfo",GetInfo)
	http.HandleFunc("/GetQuestion",GetQuestion)
	http.HandleFunc("/SetAnswer",SetAnswer)
	http.HandleFunc("/NewExam",NewExam)
	http.HandleFunc("/DelExam",DelExam)
	http.HandleFunc("/CheckStudent",CheckStudent)
	http.HandleFunc("/Upload",Upload)
	http.HandleFunc("/GetAllMark",GetAllMark)
	http.HandleFunc("/Download",Download)
	http.HandleFunc("/Cheat",Cheat)
	http.HandleFunc("/GetCheat",GetCheat)
	log.Println("考试服务监听端口: 88...")
	if err := http.ListenAndServe(":88", nil); err != nil {
		log.Fatal("启动考试服务出现问题:", err)
	}
}

func Init(){
	Paths := []string{"data/选择题/选择题.txt","data/判断题/判断题.txt","data/填空题/填空题.txt","data/简答题/简答题.txt","data/操作题/操作题.txt"}
	for i,v := range Paths{
		Data,err := ioutil.ReadFile(v)
		if err != nil{
			log.Println(err)
			os.Exit(-1)
		}
		lines := strings.Split(string(Data),"\n")
		for count:=0 ; count < len(lines) ;count ++ {
			lines[count] = strings.TrimSpace(lines[count])
			lines[count] = strings.Replace(lines[count],"\\n","",-1)
		}
		number := 0
		for count:=0 ; count < len(lines) ; {
			if i == 0{
				var question Question
				question.number = number
				question.title = lines[count]
				question.choice = lines[count+1]
				question.answer = lines[count+2]
				count = count + 3
				AllChoice = append(AllChoice,question)
			} else if i == 1 {
				var question Question
				question.number = number
				question.title = lines[count]
				question.answer = lines[count+1]
				count = count + 2
				AllJudgment = append(AllJudgment,question)
			} else if i == 2 {
				var question Question
				question.number = number
				question.title = lines[count]
				question.answer = lines[count+1]
				count = count + 2
				AllCompletion = append(AllCompletion,question)
			} else if i == 3 {
				var question Question
				question.number = number
				question.title = lines[count]
				question.answer = lines[count+1]
				count = count + 2
				AllAnswer = append(AllAnswer,question)
			} else if i == 4 {
				var question Question
				question.number = number
				question.title = lines[count]
				question.answer = lines[count+1]
				count = count + 2
				AllPractice = append(AllPractice,question)
			}
			number++
		}
	}
	log.Printf("导入选择题：%d道\n",len(AllChoice))
	log.Printf("导入填空题：%d道\n",len(AllCompletion))
	log.Printf("导入判断题：%d道\n",len(AllJudgment))
	log.Printf("导入简答题：%d道\n",len(AllAnswer))
	log.Printf("导入操作题：%d道\n",len(AllPractice))
	log.Printf("共导入题目: %d道\n",len(AllChoice)+len(AllCompletion)+len(AllJudgment)+len(AllAnswer)+len(AllPractice))

	Data,err := ioutil.ReadFile("data/学生信息/学生信息.txt")
	if err != nil{
		log.Fatal(err)
	}
	lines := strings.Split(string(Data),"\n")
	for count:=0 ; count < len(lines) ;count ++ {
		lines[count] = strings.TrimSpace(lines[count])
		lines[count] = strings.Replace(lines[count],"\\n","",-1)
	}
	for _,v := range lines{
		if v != ""{
			info := strings.Split(v," ")
			ID := info[0]
			name := info[1]
			var temp StudentInfo
			temp.ID = ID
			temp.name = name
			temp.choice.mark = "0"
			temp.judgment.mark = "0"
			temp.completion.mark = "0"
			temp.answer.mark = "0"
			temp.practice.mark = "0"
			temp.cheat1 = 0
			temp.cheat2 = 0
			temp.cheat3 = 0
			AllStudent = append(AllStudent,temp)
		}
	}

	log.Printf("共导入学生: %d名\n",len(AllStudent))
}

func GetCheat(w http.ResponseWriter, r *http.Request){
	str := `[{"name":"`+AllStudent[0].name+`","id":"`+AllStudent[0].ID+`","cheat1":`+strconv.Itoa(AllStudent[0].cheat1)+`,"cheat2":`+strconv.Itoa(AllStudent[0].cheat2)+`,"cheat3":`+strconv.Itoa(AllStudent[0].cheat3)+`}`
	for i:=1;i<len(AllStudent);i++{
		str = str + `,` + `{"name":"`+AllStudent[i].name+`","id":"`+AllStudent[i].ID+`","cheat1":`+strconv.Itoa(AllStudent[i].cheat1)+`,"cheat2":`+strconv.Itoa(AllStudent[i].cheat2)+`,"cheat3":`+strconv.Itoa(AllStudent[i].cheat3)+`}`
	}
	str = str + "]"
	w.Write([]byte(str))
	log.Println("请求获取作弊名单")
}

func Cheat(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	id := r.FormValue("id")
	CheatType := r.FormValue("type")
	for i:=0;i<len(AllStudent);i++{
		if id == AllStudent[i].ID {
			if CheatType == "0" {
				AllStudent[i].cheat1++
			} else if CheatType == "1" {
				AllStudent[i].cheat2++
			} else if CheatType == "2" {
				AllStudent[i].cheat3++
			}
			break
		}
	}
	log.Println("请求添加作弊信息，学号: ",id)
}

func Download(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	msg := r.FormValue("msg")
	if msg == "All" {
		zipDir("upload/", "data/all.zip")
		f, err := os.OpenFile("data/all.zip", os.O_RDONLY,0600)
		defer f.Close()
		if err !=nil {
			w.Write([]byte(err.Error()))
			return
		} else {
			HTMLByte,err:=ioutil.ReadAll(f)
			if err != nil{
				w.Write([]byte(err.Error()))
				return
			}
			w.Write(HTMLByte)
			log.Println("请求下载答案文件: ","data/all.zip")
		}
	}
}

func GetAllMark(w http.ResponseWriter,r *http.Request) {
	str := `[{"name":"`+ AllStudent[0].name +`","id":"`+ AllStudent[0].ID +`","choice":`+ AllStudent[0].choice.mark +`,"judgment":`+ AllStudent[0].judgment.mark +`,"completion":`+ AllStudent[0].completion.mark +`}`
	for i:=1;i<len(AllStudent);i++ {
		str = str + `,{"name":"`+ AllStudent[i].name +`","id":"`+ AllStudent[i].ID +`","choice":`+ AllStudent[i].choice.mark +`,"judgment":`+ AllStudent[i].judgment.mark +`,"completion":`+ AllStudent[i].completion.mark +`}`
	}
	str = str + "]"
	w.Write([]byte(str))
	log.Println("请求获取全部成绩")
}

func Upload(w http.ResponseWriter, r *http.Request){
	if r.Method == "GET" {
		w.Write([]byte("<html><head><title>上传</title></head>"+
		"<body><form action='#' method=\"post\" enctype=\"multipart/form-data\">"+
		"<center><h1>操作题答案上传</h1>"+"选择文件:"+
		"<input type=\"file\" name='file'  /><br/><br/>    "+
		"<label><input type=\"submit\" value=\"上传\"/></label></form></center></body></html>"))
	} else {
		//获取文件内容 要这样获取
		file, head, err := r.FormFile("file")
		if err != nil {
			log.Fatal(err)
			return
		}
		defer file.Close()
		fileSuffix := path.Ext(head.Filename)
		if fileSuffix != ".zip" {
			log.Println("请求上传文件，响应：", head.Filename ,"非法文件")
			w.Write([]byte("非法文件"))
			return
		}
		//创建文件
		fW, err := os.Create("upload/" + head.Filename)
		if err != nil {
			log.Println("请求上传文件，响应：", head.Filename ,"文件创建失败")
			w.Write([]byte("文件创建失败"))
			return
		}
		defer fW.Close()
		_, err = io.Copy(fW, file)
		if err != nil {
			log.Println("请求上传文件，响应：", head.Filename ,"文件保存失败")
			w.Write([]byte("文件保存失败"))
			return
		}
		log.Println("请求上传文件，响应：", head.Filename ,"文件上传成功")
		w.Write([]byte("文件上传成功"))
	}
}

func CheckStudent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	StudentID := r.FormValue("id")
	str := ""
	for _,v := range AllStudent {
		if v.ID == StudentID {
			str = `{"msg":"学号验证成功"}`
			w.Write([]byte(str))
			log.Println("请求验证学号，响应:",str)
			return
		}
	}
	str = `{"msg":"学号不存在，请联系老师"}`
	w.Write([]byte(str))
	log.Println("请求验证学号，响应:",str)
	return
}

func NewExam(w http.ResponseWriter, r *http.Request) {
	if examFlag {
		str := `{"msg":"添加考试失败，目前服务器已经有考试进行或即将进行..."}`
		w.Write([]byte(str))
		log.Println(`请求添加考试，响应:`,str)
		return
	}
	r.ParseForm()
	log.Println(r.FormValue("name"))
	examInfo.name = r.FormValue("name")
	starttime,_ := strconv.ParseInt(r.FormValue("start"), 10, 64)
	endtime,_ := strconv.ParseInt(r.FormValue("end"), 10, 64)
	choiceCount,_ := strconv.Atoi(r.FormValue("count0"))
	judgmentCount,_ := strconv.Atoi(r.FormValue("count1"))
	completionCount,_ := strconv.Atoi(r.FormValue("count2"))
	answerCount,_ := strconv.Atoi(r.FormValue("count3"))
	practiceCount,_ := strconv.Atoi(r.FormValue("count4"))
	examInfo.startTime = starttime
	examInfo.endTime = endtime
	examInfo.choiceCount = choiceCount
	examInfo.judgmentCount = judgmentCount
	examInfo.completionCount = completionCount
	examInfo.answerCount = answerCount
	examInfo.practiceCount = practiceCount
	go MonitorExam()
	log.Printf("成功添加考试:%s 开始Unix时间: %d 结束Unix时间: %d 选择题: %d道,填空题: %d道,判断题: %d道,简答题: %d道,操作题: %d道",examInfo.name,examInfo.startTime,examInfo.endTime,examInfo.choiceCount,examInfo.completionCount,examInfo.judgmentCount,examInfo.answerCount,examInfo.practiceCount)
	for i:=0;i<len(AllStudent);i++{
		AllStudent[i].choice.origin = generateRandomNumber(0,len(AllChoice),examInfo.choiceCount)
		AllStudent[i].choice.json = GetQuestionJSON("0",AllStudent[i].choice.origin)
		AllStudent[i].judgment.origin = generateRandomNumber(0,len(AllJudgment),examInfo.judgmentCount)
		AllStudent[i].judgment.json = GetQuestionJSON("1",AllStudent[i].judgment.origin)
		AllStudent[i].completion.origin = generateRandomNumber(0,len(AllCompletion),examInfo.completionCount)
		AllStudent[i].completion.json = GetQuestionJSON("2",AllStudent[i].completion.origin)
		AllStudent[i].answer.origin = generateRandomNumber(0,len(AllAnswer),examInfo.answerCount)
		AllStudent[i].answer.json = GetQuestionJSON("3",AllStudent[i].answer.origin)
		AllStudent[i].practice.origin = generateRandomNumber(0,len(AllPractice),examInfo.practiceCount)
		AllStudent[i].practice.json = GetQuestionJSON("4",AllStudent[i].practice.origin)
	}
	w.Write([]byte(`{"msg":"success"}`))
	log.Println(`请求添加考试，响应:{"msg":"success"}`)
	examFlag = true
}

func DelExam(w http.ResponseWriter, r *http.Request) {
	examInfo.name = ""
	examInfo.startTime = 0
	examInfo.endTime = 0
	examInfo.choiceCount = 0
	examInfo.judgmentCount = 0
	examInfo.completionCount = 0
	examInfo.answerCount = 0
	examInfo.practiceCount = 0
	examFlag = false
	w.Write([]byte(`{"msg":"success"}`))
	log.Println(`请求删除考试，响应:{"msg":"success"}`)
}

func MonitorExam(){
	log.Println("启动考试状态监控线程")
	for ;;{
		if time.Now().Unix() > examInfo.startTime{
			if time.Now().Unix() > examInfo.endTime {
				examInfo.state = "考试结束"
				examFlag = false
				log.Println("考试结束，考试监控线程退出")
				return
			} else {
				examInfo.state = "考试开始"
			}
		} else if time.Now().Unix() < examInfo.startTime {
			examInfo.state = "考试未开始"
		}
		if examFlag == false {
			log.Println("考试被删除，考试监控线程退出")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func GetInfo(w http.ResponseWriter, r *http.Request) {
	if examFlag {
		str := `{"name":"`+ examInfo.name +`","state":"`+ examInfo.state +`"}`
		w.Write([]byte(str))
		log.Println("请求获取考试信息，响应:",str)
	} else {
		str := `{"name":"当前无考试或连接服务器失败","state":"无考试"}`
		w.Write([]byte(str))
		log.Println("请求获取考试信息，响应:",str)
	}
}

func GetQuestionJSON(QuestionType string,QuestionList []int) string{
	if QuestionType == "0"{
		temp := "["
		for i,v := range QuestionList{
			temp = temp + `{"number":`+ strconv.Itoa(AllChoice[v].number) +`,"title":"`+ AllChoice[v].title +`","choice":"`+ AllChoice[v].choice +`","answer":"`+ AllChoice[v].answer +`"}`
			if i != len(QuestionList)-1{
				temp = temp + ","
			}
		}
		temp = temp + "]"
		temp = strings.Replace(temp,"\\n","",-1)
		return temp
	} else if QuestionType == "1" {
		temp := "["
		for i,v := range QuestionList{
			temp = temp + `{"number":`+ strconv.Itoa(AllJudgment[v].number) +`,"title":"`+ AllJudgment[v].title +`","choice":"`+ AllJudgment[v].choice +`","answer":"`+ AllJudgment[v].answer +`"}`
			if i != len(QuestionList)-1{
				temp = temp + ","
			}
		}
		temp = temp + "]"
		temp = strings.Replace(temp,"\\n","",-1)
		return temp
	} else if QuestionType == "2" {
		temp := "["
		for i,v := range QuestionList{
			temp = temp + `{"number":`+ strconv.Itoa(AllCompletion[v].number) +`,"title":"`+ AllCompletion[v].title +`","choice":"`+ AllCompletion[v].choice +`","answer":"`+ AllCompletion[v].answer +`"}`
			if i != len(QuestionList)-1{
				temp = temp + ","
			}
		}
		temp = temp + "]"
		temp = strings.Replace(temp,"\\n","",-1)
		return temp
	} else if QuestionType == "3" {
		temp := "["
		for i,v := range QuestionList{
			temp = temp + `{"number":`+ strconv.Itoa(AllAnswer[v].number) +`,"title":"`+ AllAnswer[v].title +`","choice":"`+ AllAnswer[v].choice +`","answer":"`+ AllAnswer[v].answer +`"}`
			if i != len(QuestionList)-1{
				temp = temp + ","
			}
		}
		temp = temp + "]"
		temp = strings.Replace(temp,"\\n","",-1)
		return temp
	} else if QuestionType == "4" {
		temp := "["
		for i,v := range QuestionList{
			temp = temp + `{"number":`+ strconv.Itoa(AllPractice[v].number) +`,"title":"`+ AllPractice[v].title +`","choice":"`+ AllPractice[v].choice +`","answer":"`+ AllPractice[v].answer +`"}`
			if i != len(QuestionList)-1{
				temp = temp + ","
			}
		}
		temp = temp + "]"
		temp = strings.Replace(temp,"\\n","",-1)
		return temp
	}
	return ""
}

func SetAnswer(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	StudentID := r.FormValue("id")
	answerStr := r.FormValue("answer")
	answerType := r.FormValue("type")
	mark := r.FormValue("mark")
	for i:=0 ;i<len(AllStudent);i++{
		if AllStudent[i].ID == StudentID {
			temp := strings.Split(answerStr,"|@|")
			if answerType == "0" {
				for _,v := range temp {
					AllStudent[i].choice.answer = append(AllStudent[i].choice.answer,v)
				}
				AllStudent[i].choice.mark = mark
			} else if answerType == "1" {
				for _,v := range temp {
					AllStudent[i].judgment.answer = append(AllStudent[i].judgment.answer,v)
				}
				AllStudent[i].judgment.mark = mark
			} else if answerType == "2" {
				for _,v := range temp {
					AllStudent[i].completion.answer = append(AllStudent[i].completion.answer,v)
				}
				AllStudent[i].completion.mark = mark
			} else if answerType == "3" {
				for _,v := range temp {
					AllStudent[i].answer.answer = append(AllStudent[i].answer.answer,v)
				}
				AllStudent[i].answer.mark = mark
			} else if answerType == "4" {
				for _,v := range temp {
					AllStudent[i].practice.answer = append(AllStudent[i].practice.answer,v)
				}
				AllStudent[i].practice.mark = mark
			}
			break
		}
	}
	log.Println("请求上传成绩，用户: ",StudentID)
}

func GetQuestion(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	QuestionType := r.FormValue("type")
	StudentID := r.FormValue("id")
	for _,v := range AllStudent {
		if v.ID == StudentID {
			if QuestionType == "0" {
				w.Write([]byte(v.choice.json))
			} else if QuestionType == "1" {
				w.Write([]byte(v.judgment.json))
			} else if QuestionType == "2" {
				w.Write([]byte(v.completion.json))
			} else if QuestionType == "3" {
				w.Write([]byte(v.answer.json))
			} else if QuestionType == "4" {
				w.Write([]byte(v.practice.json))
			}
			break
		}
	}
	log.Println("请求获取试题，学生:",StudentID)
}

func connectTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("success"))
	log.Println("请求测试连接服务器")
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("访问方式错误..."))
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	defaultRoot := "data/操作题/"
	r.ParseForm()
	filename := r.FormValue("file")
	if filename == ""{
    	w.Write([]byte("please input file name"))
    	return
  	}
  	f, err := os.OpenFile(defaultRoot+filename, os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(HTMLByte)
		log.Println("请求下载文件: ",filename)
	}
}

func generateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}
	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

func zipDir(dir, zipFile string) {

    fz, err := os.Create(zipFile)
    if err != nil {
        log.Fatalf("Create zip file failed: %s\n", err.Error())
    }
    defer fz.Close()

    w := zip.NewWriter(fz)
    defer w.Close()

    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            fDest, err := w.Create(path[len(dir)+1:])
            if err != nil {
                log.Printf("Create failed: %s\n", err.Error())
                return nil
            }
            fSrc, err := os.Open(path)
            if err != nil {
                log.Printf("Open failed: %s\n", err.Error())
                return nil
            }
            defer fSrc.Close()
            _, err = io.Copy(fDest, fSrc)
            if err != nil {
                log.Printf("Copy failed: %s\n", err.Error())
                return nil
            }
        }
        return nil
    })
}