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

실행 결과는 `logs/YYYY-MM-DD_HHmmss/` 폴더에 서버별로 저장됩니다.
스케줄 실행 시 `logs/{스케줄이름}/YYYY-MM-DD_HHmmss/` 하위에 저장됩니다.

## 변경 이력

### v1.1
- 스케줄 완료 시 이메일 알림 기능 추가 (SMTP 설정 암호화 저장)
- 이메일 알림 HTML 템플릿 및 상태 배지 적용
- alert() 다이얼로그를 비동기 토스트 알림으로 전면 교체
- 스케줄 실행 큐 시스템 추가 (동시 실행 시 스케줄 누락 방지)
- 동일 이름 스케줄 중복 생성 방지
- 같은 날 재실행 시 로그 덮어쓰기 방지 (타임스탬프 포함 디렉토리)

### v1.0.0
- 초기 릴리스
- 다중 Cisco 장비 SSH 접속 및 명령어 일괄 실행
- 실시간 로그 스트리밍
- 스케줄 실행 기능
