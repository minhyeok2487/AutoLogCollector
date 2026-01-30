# 스케줄링 완전 가이드

[← 목차로 돌아가기](./USER_GUIDE.md)

---

## 스케줄 생성

1. 좌측 메뉴에서 **Schedule** 클릭
2. **+ New Schedule** 버튼 클릭
3. 스케줄 양식 작성

### 기본 정보

- **Name**: 스케줄 이름 (중복 불가)

### 인증 정보

- **Username / Password**: 스케줄 실행에 사용할 SSH 계정
- **Enable Password**: Enable Mode 사용 시 비밀번호

### 스케줄 타입

| 타입 | 설정 항목 | 예시 |
|------|-----------|------|
| Daily | 실행 시각 | 매일 09:00 |
| Weekly | 요일 선택 + 실행 시각 | 월~금 18:00 |
| Monthly | 날짜 + 실행 시각 | 매월 1일 00:00 |

### 실행 옵션

- **Timeout**: 명령 응답 대기 시간
- **Disable Paging**: 페이징 비활성화
- **Enable Mode**: 특권 모드 진입
- **Auto Export Excel**: 자동 Excel 생성

### 대상 서버 및 명령어

- 직접 서버와 명령어를 입력하거나
- **Copy from Execution** 버튼으로 현재 Execution 화면의 서버 목록과 명령어를 복사

---

## Execution에서 서버/명령어 복사

스케줄 생성 화면에서 **Copy from Execution** 기능을 사용하면 Execution 화면에 현재 설정된 서버 목록과 명령어를 그대로 스케줄에 복사할 수 있습니다. 매번 동일한 서버와 명령어를 다시 입력할 필요가 없습니다.

---

## 이메일 알림 설정

### 1단계: SMTP 설정

1. 좌측 하단 **Settings** → **SMTP Settings** 클릭
2. SMTP 서버 정보 입력:

| 항목 | 예시 |
|------|------|
| SMTP Server | smtp.gmail.com |
| Port | 587 |
| Username | user@gmail.com |
| Password | 앱 비밀번호 |

3. **Save** 클릭 (인증 정보는 AES-256-GCM으로 암호화 저장)

### 2단계: 스케줄에 수신자 지정

1. 스케줄 생성/수정 화면에서 **Email Notification** 활성화
2. **Email To** 필드에 수신자 이메일 주소 입력
3. 스케줄 저장

### 발송되는 이메일 내용

- HTML 형식의 실행 결과 요약
- 서버별 성공/실패 상태 배지
- 로그 파일 및 Excel이 포함된 ZIP 파일 첨부

---

## 스케줄 로그 저장 경로

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

[← 고급 기능](./03-advanced.md) | [다음: 설정 파일 레퍼런스 →](./05-config-reference.md)
