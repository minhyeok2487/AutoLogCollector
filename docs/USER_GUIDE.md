# AutoLogCollector 사용자 가이드

## 목차

1. [빠른 시작 가이드](#빠른-시작-가이드)
2. [화면별 상세 가이드](#화면별-상세-가이드)
3. [고급 기능](#고급-기능)
4. [스케줄링 완전 가이드](#스케줄링-완전-가이드)
5. [설정 파일 레퍼런스](#설정-파일-레퍼런스)
6. [FAQ / 트러블슈팅](#faq--트러블슈팅)

---

## 빠른 시작 가이드

처음 사용하는 경우 아래 단계를 따라 진행하세요.

1. `AutoLogCollector.exe` 실행
2. **Username**과 **Password** 입력 (장비 SSH 접속 계정)
3. **+ Add** 버튼으로 대상 서버 추가 (IP, Hostname)
4. 명령어 입력 영역에 실행할 명령어 작성 (한 줄에 하나씩)
5. **Run Execution** 클릭
6. **Live Logs** 탭에서 실시간 로그 확인
7. **Results** 탭에서 실행 결과 확인 및 Excel 내보내기

---

## 화면별 상세 가이드

### Execution 화면

메인 화면으로, 접속 정보 설정부터 실행까지 모든 작업을 수행합니다.

![Execution 화면](../01.png)

#### Connection Settings

| 항목 | 설명 | 기본값 |
|------|------|--------|
| Username | SSH 접속 계정 | - |
| Password | SSH 접속 비밀번호 | - |
| Timeout | 접속 및 명령 응답 대기 시간 (초) | 10 |
| Disable Paging | `terminal length 0` 자동 전송 | 활성화 |
| Auto Export Excel | 완료 시 자동 Excel 생성 | 비활성화 |
| Enable Mode | 특권 모드 진입 | 비활성화 |
| Enable Password | Enable Mode 비밀번호 (Enable Mode 활성화 시 표시) | - |

#### Target Servers

서버 목록을 관리하는 영역입니다.

- **Import CSV**: `ip,hostname` 형식의 CSV 파일에서 서버 목록을 일괄 불러옵니다.
- **Export CSV**: 현재 서버 목록을 CSV 파일로 내보냅니다.
- **+ Add**: 테이블에 새 행을 추가하여 IP와 Hostname을 직접 입력합니다.
- **삭제**: 각 행의 삭제 버튼으로 개별 서버를 제거합니다.
- **서버별 인증(🔒)**: 잠금 아이콘을 클릭하면 해당 서버에만 적용되는 별도 인증 정보를 설정할 수 있습니다.

#### Commands

- 텍스트 영역에 명령어를 한 줄씩 입력합니다.
- **Import TXT**: 텍스트 파일에서 명령어를 불러옵니다.
- **Export TXT**: 현재 명령어를 텍스트 파일로 저장합니다.
- 화면 하단에 입력된 명령어 수가 표시됩니다.

#### 실행

- **Run Execution**: 설정된 서버 목록에 순차적으로 SSH 접속하여 명령어를 실행합니다.
- **Stop**: 실행 중 중단합니다. 이미 진행 중인 서버 작업은 완료된 후 중단됩니다.
- 진행률 바와 완료 서버 수가 실시간으로 표시됩니다.

---

### Results 화면

![Results 화면](../02.png)

실행이 완료되면 서버별 결과를 확인할 수 있습니다.

| 열 | 설명 |
|----|------|
| Hostname | 서버 호스트명 |
| IP | 서버 IP 주소 |
| Status | 성공(Success) 또는 실패(Failed) |
| Duration | 명령 실행 소요 시간 |
| Action | View Log 버튼 |

- **View Log**: 해당 서버의 전체 로그를 모달 창에서 확인합니다.
- **Export Excel**: 명령어별로 시트가 구분된 Excel 파일을 생성합니다. 각 열은 서버, 각 행은 해당 명령의 출력입니다.
- **Open Logs Folder**: Windows 파일 탐색기에서 로그 폴더를 엽니다.

상단 요약 바에서 성공/실패/전체 서버 수를 한눈에 확인할 수 있습니다.

---

### Live Logs 화면

실행 중 모든 서버의 SSH 세션 출력을 실시간으로 확인합니다.

- **Combined 탭**: 모든 서버의 로그가 타임스탬프와 서버명과 함께 통합 표시됩니다.
- **서버별 탭**: 실행이 시작되면 각 서버 이름으로 탭이 자동 생성됩니다. 탭을 클릭하면 해당 서버의 로그만 필터링됩니다.
- **Auto-scroll**: 활성화하면 새 로그 수신 시 자동으로 하단으로 스크롤합니다.
- **Clear**: 화면의 로그를 초기화합니다 (파일에 저장된 로그는 유지).

---

### Schedule 화면

등록된 스케줄 목록을 관리합니다.

| 열 | 설명 |
|----|------|
| Name | 스케줄 이름 |
| Type | Daily / Weekly / Monthly |
| Time | 실행 시각 |
| Next Run | 다음 실행 예정 시각 |
| Last Run | 마지막 실행 시각 |
| Status | Enabled / Disabled |
| Actions | 편집, 삭제, 활성화/비활성화 |

자세한 내용은 [스케줄링 완전 가이드](#스케줄링-완전-가이드)를 참조하세요.

---

## 고급 기능

### 병렬 실행 (Concurrent) 설정

Concurrent 값은 동시에 SSH 접속을 시도하는 서버 수를 제어합니다.

- 서버 수가 적은 경우(10대 이하): 전체 서버 수와 동일하게 설정
- 서버 수가 많은 경우(10대 이상): 5~10 정도로 설정하여 네트워크 부하 분산
- 너무 높은 값은 네트워크 병목이나 장비 부하를 유발할 수 있습니다

### 서버별 개별 인증 (Per-server Credentials)

환경에 따라 서버마다 다른 계정이 필요한 경우 사용합니다.

1. Target Servers 테이블에서 해당 서버의 **🔒** 아이콘 클릭
2. 팝업에서 Username, Password, Enable Password 입력
3. **Save** 클릭

설정된 서버는 전역 인증 정보 대신 개별 인증 정보를 사용합니다. 개별 인증을 제거하면 전역 인증으로 돌아갑니다.

인증 정보는 `config/servers.json`에 AES-256-GCM으로 암호화되어 저장됩니다.

### Enable Mode 사용법

`show running-config` 등 특권 모드가 필요한 명령을 실행할 때 사용합니다.

1. Connection Settings에서 **Enable Mode** 체크박스 활성화
2. **Enable Password** 입력 (로그인 비밀번호와 동일한 경우 비워둘 수 있음)
3. 실행 시 자동으로 `enable` 명령을 전송하고 비밀번호를 입력합니다

### Disable Paging 동작 원리

Cisco 장비는 기본적으로 출력이 길면 `--More--` 프롬프트로 페이징합니다. Disable Paging을 활성화하면:

1. 세션 시작 시 `terminal length 0` 명령을 자동 전송
2. 이후 모든 명령의 출력이 끊기지 않고 전체 표시
3. 만약 `--More--` 프롬프트가 감지되면 자동으로 스페이스를 전송하여 계속 진행

### Auto Export Excel

활성화하면 모든 서버의 실행이 완료된 직후 자동으로 `results.xlsx` 파일이 로그 폴더에 생성됩니다. 수동으로 Results 화면에서 **Export Excel** 버튼을 클릭할 필요가 없습니다.

---

## 스케줄링 완전 가이드

### 스케줄 생성

1. 좌측 메뉴에서 **Schedule** 클릭
2. **+ New Schedule** 버튼 클릭
3. 스케줄 양식 작성:

#### 기본 정보
- **Name**: 스케줄 이름 (중복 불가)

#### 인증 정보
- **Username / Password**: 스케줄 실행에 사용할 SSH 계정
- **Enable Password**: Enable Mode 사용 시 비밀번호

#### 스케줄 타입

| 타입 | 설정 항목 | 예시 |
|------|-----------|------|
| Daily | 실행 시각 | 매일 09:00 |
| Weekly | 요일 선택 + 실행 시각 | 월~금 18:00 |
| Monthly | 날짜 + 실행 시각 | 매월 1일 00:00 |

#### 실행 옵션
- **Timeout**: 명령 응답 대기 시간
- **Disable Paging**: 페이징 비활성화
- **Enable Mode**: 특권 모드 진입
- **Auto Export Excel**: 자동 Excel 생성

#### 대상 서버 및 명령어
- 직접 서버와 명령어를 입력하거나
- **Copy from Execution** 버튼으로 현재 Execution 화면의 서버 목록과 명령어를 복사

### Execution에서 서버/명령어 복사

스케줄 생성 화면에서 **Copy from Execution** 기능을 사용하면 Execution 화면에 현재 설정된 서버 목록과 명령어를 그대로 스케줄에 복사할 수 있습니다. 매번 동일한 서버와 명령어를 다시 입력할 필요가 없습니다.

### 이메일 알림 설정

#### 1단계: SMTP 설정

1. 좌측 하단 **Settings** → **SMTP Settings** 클릭
2. SMTP 서버 정보 입력:

| 항목 | 예시 |
|------|------|
| SMTP Server | smtp.gmail.com |
| Port | 587 |
| Username | user@gmail.com |
| Password | 앱 비밀번호 |

3. **Save** 클릭 (인증 정보는 AES-256-GCM으로 암호화 저장)

#### 2단계: 스케줄에 수신자 지정

1. 스케줄 생성/수정 화면에서 **Email Notification** 활성화
2. **Email To** 필드에 수신자 이메일 주소 입력
3. 스케줄 저장

#### 발송되는 이메일 내용

- HTML 형식의 실행 결과 요약
- 서버별 성공/실패 상태 배지
- 로그 파일 및 Excel이 포함된 ZIP 파일 첨부

### 스케줄 로그 저장 경로

스케줄 실행 로그는 다음 경로에 저장됩니다:

```
logs/{스케줄이름}/YYYY-MM-DD_HHmmss/
├── Router1.log
├── Switch1.log
├── ...
└── results.xlsx (Auto Export Excel 활성화 시)
```

같은 날 여러 번 실행되어도 타임스탬프(`HHmmss`)로 구분되어 덮어쓰기가 발생하지 않습니다.

---

## 설정 파일 레퍼런스

### servers.csv

서버 목록을 일괄 가져오기/내보내기할 때 사용하는 CSV 파일입니다.

```csv
ip,hostname
192.168.0.1,Router1
192.168.0.2,Switch1
10.0.0.1,CoreSwitch
```

- 첫 번째 행은 헤더(`ip,hostname`)
- 인증 정보는 포함되지 않음 (보안)

### commands.txt

실행할 명령어 목록을 가져오기/내보내기할 때 사용하는 텍스트 파일입니다.

```
show version
show ip interface brief
show processes cpu | include CPU
show memory statistics
show running-config
exit
```

- 한 줄에 하나의 명령어
- `exit`는 선택사항 (자동 처리됨)
- 파이프(`|`) 포함 명령어 사용 가능

### credentials.json

전역 인증 정보를 저장하는 JSON 파일입니다.

```json
{
  "user": "admin",
  "password": "password"
}
```

### config/servers.json (자동 생성)

UI에서 서버별 개별 인증을 설정하면 자동으로 생성됩니다. 인증 정보는 AES-256-GCM으로 암호화됩니다.

```json
[
  {
    "ip": "192.168.0.1",
    "hostname": "Router1",
    "username": "admin",
    "password": "(암호화된 문자열)",
    "enablePassword": "(암호화된 문자열)"
  }
]
```

### config/schedules.json (자동 생성)

스케줄 정보가 자동 저장됩니다. 직접 편집하지 않는 것을 권장합니다.

### config/smtp.json (자동 생성)

SMTP 설정이 암호화되어 저장됩니다.

```json
{
  "server": "smtp.gmail.com",
  "port": 587,
  "username": "user@gmail.com",
  "password": "(암호화된 문자열)"
}
```

### config/encryption.key (자동 생성)

AES-256-GCM 암호화에 사용되는 키 파일입니다. 설치별로 고유하게 생성되며, 삭제 시 저장된 모든 암호화 정보를 읽을 수 없게 됩니다.

---

## FAQ / 트러블슈팅

### Windows SmartScreen 경고가 나타납니다

"추가 정보" → "실행"을 클릭하세요. 코드 서명이 없는 실행 파일에 대해 Windows가 표시하는 기본 경고입니다. 보안 위험은 없습니다.

### SSH 접속이 실패합니다

다음 항목을 순서대로 확인하세요:

1. **네트워크 연결**: 해당 장비 IP로 ping 확인
2. **SSH 포트**: 장비의 22번 포트가 열려 있는지 확인 (`telnet {IP} 22`)
3. **방화벽**: 로컬 PC와 장비 사이의 방화벽 규칙 확인
4. **인증 정보**: Username과 Password가 정확한지 확인
5. **SSH 서비스**: 장비에서 SSH 서비스가 활성화되어 있는지 확인
6. **서버별 인증**: 해당 서버에 개별 인증이 설정되어 있다면 해당 정보도 확인

### Timeout 조정 가이드

응답이 느린 장비에서 명령 출력이 잘리거나 접속이 끊기는 경우:

- **기본값 (10초)**: 일반적인 Cisco 장비에 적합
- **20~30초**: `show running-config` 등 출력이 많은 명령 실행 시
- **30~60초**: 부하가 높거나 응답이 느린 장비, 또는 WAN 구간 장비

Timeout은 Connection Settings에서 1~60초 범위로 설정 가능합니다.

### 명령 실행 중 출력이 잘립니다

- **Disable Paging**이 활성화되어 있는지 확인하세요.
- 출력이 매우 긴 경우 **Timeout** 값을 늘려보세요.
- `--More--` 프롬프트가 자동으로 처리되지 않는 경우, 명령어 목록 맨 위에 `terminal length 0`을 수동으로 추가해보세요.

### 로그 파일 위치 및 구조

```
logs/
├── 2026-01-27_090000/          # 수동 실행 로그
│   ├── Router1.log
│   ├── Switch1.log
│   └── results.xlsx
├── DailyBackup/                # 스케줄 실행 로그
│   ├── 2026-01-27_090000/
│   │   ├── Router1.log
│   │   └── results.xlsx
│   └── 2026-01-28_090000/
│       └── ...
```

- 수동 실행: `logs/YYYY-MM-DD_HHmmss/`
- 스케줄 실행: `logs/{스케줄이름}/YYYY-MM-DD_HHmmss/`
- 각 서버별로 `{Hostname}.log` 파일 생성
- Auto Export Excel 활성화 시 `results.xlsx` 포함

### Excel 파일이 생성되지 않습니다

- **Auto Export Excel**을 활성화했는지 확인하세요.
- 수동 생성: Results 화면에서 **Export Excel** 버튼 클릭
- 실행이 정상 완료되지 않은 경우(모든 서버 실패 등) Excel이 생성되지 않을 수 있습니다.

### 스케줄이 실행되지 않습니다

- **앱이 실행 중**이어야 스케줄이 동작합니다. 앱을 종료하면 스케줄도 중단됩니다.
- Schedule 화면에서 해당 스케줄이 **Enabled** 상태인지 확인하세요.
- **Next Run** 시각이 올바른지 확인하세요.
- 동일 시각에 여러 스케줄이 설정된 경우, 실행 큐 시스템에 의해 순차적으로 실행됩니다.

### 이메일 발송이 실패합니다

SMTP 설정을 점검하세요:

1. **SMTP 서버 주소**: 올바른 서버 주소 확인 (예: `smtp.gmail.com`, `smtp.office365.com`)
2. **포트**: TLS/STARTTLS 포트 사용 (일반적으로 587)
3. **인증 정보**: SMTP 계정과 비밀번호 확인
4. **Gmail 사용 시**: 일반 비밀번호 대신 [앱 비밀번호](https://myaccount.google.com/apppasswords) 사용 필요
5. **네트워크**: SMTP 서버로의 아웃바운드 연결이 방화벽에 의해 차단되지 않는지 확인
6. **수신자 주소**: 스케줄 설정의 Email To 필드에 올바른 이메일 주소가 입력되어 있는지 확인
