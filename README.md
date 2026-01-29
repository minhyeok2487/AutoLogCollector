# AutoLogCollector

Cisco 및 네트워크 장비에 SSH로 접속하여 명령을 자동 실행하고 로그를 수집하는 GUI 도구

![Execution](01.png)
![Results](02.png)
![Log Files](03.png)

## 기능

- 여러 네트워크 장비에 동시 접속 (병렬 처리)
- 명령어 일괄 실행 및 실시간 로그 스트리밍
- 서버별 개별 인증 지원 (Per-server Credentials)
- Enable Mode / Disable Paging 자동 처리
- 실행 결과 Excel 내보내기
- 스케줄 실행 (Daily / Weekly / Monthly)
- 스케줄 완료 시 이메일 알림 (SMTP)
- 자동 업데이트

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

## 상세 사용법

> 더 자세한 가이드는 [사용자 가이드](docs/USER_GUIDE.md)를 참조하세요.

### Connection Settings

| 항목 | 설명 |
|------|------|
| Username / Password | 장비 접속에 사용할 전역 인증 정보 |
| Timeout | SSH 접속 및 명령 응답 대기 시간 (1~60초, 기본 10초) |
| Disable Paging | `terminal length 0` 자동 전송하여 페이징 방지 |
| Enable Mode | 특권 모드(`enable`) 진입 활성화. 별도 Enable Password 입력 가능 |
| Auto Export Excel | 실행 완료 시 자동으로 Excel 파일 생성 |

### Target Servers

- **Import CSV**: `ip,hostname` 형식의 CSV 파일을 불러와 서버 목록 일괄 등록
- **Export CSV**: 현재 서버 목록을 CSV로 내보내기
- **+ Add**: 수동으로 서버 추가
- **서버별 개별 인증**: 각 서버의 잠금 아이콘(🔒)을 클릭하여 해당 서버만의 Username, Password, Enable Password 설정 가능. 설정하지 않으면 전역 인증 정보 사용

### Commands

- 텍스트 영역에 직접 명령어 입력 (한 줄에 하나씩)
- **Import TXT**: 명령어 목록 파일 불러오기
- **Export TXT**: 현재 명령어 목록을 파일로 내보내기

### 실행 및 진행률 확인

1. 서버 목록과 명령어를 설정한 뒤 **Run Execution** 클릭
2. 진행률 바에서 전체 진행 상황 확인
3. 실행 중 **Stop** 버튼으로 중단 가능

### Results

- 실행 완료 후 서버별 성공/실패 상태, 소요 시간 확인
- **View Log**: 각 서버의 전체 로그를 모달로 확인
- **Export Excel**: 명령어별 시트로 구분된 Excel 파일 생성
- **Open Logs Folder**: 로그 저장 폴더를 파일 탐색기에서 열기

### Live Logs

- **Combined View**: 모든 서버의 로그를 통합하여 실시간 확인
- **서버별 탭 필터링**: 특정 서버의 로그만 선택하여 확인
- **Auto-scroll**: 새 로그 수신 시 자동 스크롤
- **Clear**: 로그 화면 초기화

## 스케줄 기능

### 스케줄 생성

Schedule 화면에서 **+ New Schedule** 버튼으로 생성합니다.

| 타입 | 설명 |
|------|------|
| Daily | 매일 지정한 시각에 실행 |
| Weekly | 선택한 요일에 지정 시각 실행 |
| Monthly | 매월 지정한 날짜·시각에 실행 |

- Execution 화면의 서버 목록과 명령어를 스케줄로 복사 가능
- 스케줄별 독립적인 인증 정보, Timeout, Disable Paging, Enable Mode 설정
- 동일 이름의 스케줄 중복 생성 방지
- 실행 큐 시스템으로 동시 실행 시 누락 방지

### 이메일 알림

1. **Settings → SMTP Settings**에서 SMTP 서버 정보 설정 (암호화 저장)
2. 스케줄 생성/수정 시 **Email Notification** 활성화 후 수신자 이메일 입력
3. 스케줄 완료 시 실행 결과 요약 및 로그 ZIP 파일이 첨부된 HTML 이메일 발송

## 설정 파일

### config/servers.csv

```csv
ip,hostname
192.168.0.1,Router1
192.168.0.2,Switch1
```

### config/commands.txt

```
show version
show ip interface brief
show processes cpu | include CPU
exit
```

### credentials.json (전역 인증)

```json
{
  "user": "admin",
  "password": "password"
}
```

### 서버별 개별 인증

서버별 인증 정보는 UI에서 설정하면 `config/servers.json`에 AES-256-GCM으로 암호화되어 자동 저장됩니다.

```json
[
  {
    "ip": "192.168.0.1",
    "hostname": "Router1",
    "username": "admin",
    "password": "(AES-256-GCM 암호화됨)",
    "enablePassword": "(AES-256-GCM 암호화됨)"
  }
]
```

## 로그

실행 결과는 `logs/YYYY-MM-DD_HHmmss/` 폴더에 서버별로 저장됩니다.
스케줄 실행 시 `logs/{스케줄이름}/YYYY-MM-DD_HHmmss/` 하위에 저장됩니다.

## FAQ / 트러블슈팅

**Q: Windows SmartScreen 경고가 나타납니다.**
A: "추가 정보" → "실행"을 클릭하세요. 서명되지 않은 실행 파일에 대한 기본 경고입니다.

**Q: SSH 접속이 실패합니다.**
A: 다음을 확인하세요:
- 장비 IP 및 SSH 포트(22) 접근 가능 여부
- 방화벽 규칙
- Username / Password 정확성
- 장비의 SSH 서비스 활성화 여부

**Q: 명령 실행 중 출력이 잘립니다.**
A: **Disable Paging**을 활성화하세요. 응답이 느린 장비는 **Timeout**을 늘려보세요.

**Q: Excel 파일이 생성되지 않습니다.**
A: Results 화면에서 **Export Excel** 버튼을 클릭하거나, **Auto Export Excel** 옵션을 활성화하세요.

**Q: 스케줄이 실행되지 않습니다.**
A: 앱이 실행 중인 상태에서만 스케줄이 동작합니다. 스케줄이 **Enabled** 상태인지 확인하세요.

**Q: 이메일 발송이 실패합니다.**
A: Settings → SMTP Settings에서 서버 주소, 포트, 인증 정보를 확인하세요. TLS/STARTTLS 포트(587)를 사용하는지 확인하세요.

## 변경 이력

### v1.1.2
- README.md 보강 (스크린샷, 상세 사용법, 스케줄 기능, FAQ 추가)
- 사용자 가이드 문서 추가 (docs/USER_GUIDE.md)

### v1.1.1
- 버전 업데이트 시 릴리스 노트 마크다운 문법 제거
- 앱 내 버전 표시 수정

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
