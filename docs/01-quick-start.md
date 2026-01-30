# 빠른 시작 가이드

[← 목차로 돌아가기](./USER_GUIDE.md)

---

처음 사용하는 경우 아래 단계를 따라 진행하세요.

## 1단계: 실행

`AutoLogCollector.exe`를 실행합니다.

> Windows SmartScreen 경고가 나타나면 "추가 정보" → "실행"을 클릭하세요.

## 2단계: 인증 정보 입력

**Username**과 **Password**를 입력합니다. 대상 장비의 SSH 접속 계정입니다.

## 3단계: 대상 서버 추가

**+ Add** 버튼으로 서버를 추가하거나, **Import CSV** 버튼으로 서버 목록 파일을 불러옵니다.

CSV 파일 형식:
```csv
ip,hostname
192.168.0.1,Router1
192.168.0.2,Switch1
```

## 4단계: 명령어 입력

명령어 입력 영역에 실행할 명령어를 한 줄에 하나씩 작성합니다.

```
show version
show ip interface brief
show running-config
```

## 5단계: 실행

**Run Execution** 버튼을 클릭합니다.

## 6단계: 로그 확인

**Live Logs** 탭에서 실시간 로그를 확인합니다.

## 7단계: 결과 확인

**Results** 탭에서 실행 결과를 확인하고, 필요하면 **Export Excel** 버튼으로 Excel 파일을 생성합니다.

---

[다음: 화면별 상세 가이드 →](./02-screens.md)
