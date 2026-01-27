# Cisco Automation Tool

Cisco 장비에 SSH로 접속하여 명령을 자동 실행하는 GUI 도구

## 기능

- 여러 Cisco 장비에 동시 접속 (병렬 처리)
- 명령어 일괄 실행
- 실시간 로그 스트리밍
- 서버별 탭으로 로그 확인
- 실행 결과 로그 파일 저장

## 요구사항

- Go 1.21+
- Wails v2

## 설치

```bash
# Wails 설치
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 의존성 설치
go mod tidy
```

## 실행

```bash
# 개발 모드
wails dev

# 빌드
wails build
```

## 설정 파일

### config/servers.csv
```csv
ip,hostname
192.168.0.1,Router1
192.168.0.2,Switch1
```

### config/commands.txt
```
terminal length 0
show version
show ip interface brief
exit
```

### credentials.json
```json
{
  "user": "admin",
  "password": "password"
}
```

## 사용법

1. Username / Password 입력
2. Concurrent 설정 (동시 접속 수)
3. Servers File 선택 (CSV)
4. Commands File 선택 (TXT)
5. Run 클릭
6. Live Logs 탭에서 실시간 로그 확인

## 로그

실행 결과는 `logs/YYYY-MM-DD/` 폴더에 서버별로 저장됩니다.
