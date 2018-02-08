FROM python:3.6-slim

WORKDIR /bashpush

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY bashpush.py .
COPY run.sh .

CMD ["./run.sh"]
